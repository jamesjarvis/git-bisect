// Package dag implements directed acyclic graphs (DAGs).
package dag

import (
	"fmt"
	"log"
	"math"
	"runtime"
	"sync"
)

/*
This is a guesstimate of what we have
	A
   / \
  B   C
 / \ / \
D   E   F
 \ /
  G

Should be:

Vertices:
A
B
C
D
E
F

Edges:
B -> A
C -> A
D -> B
E -> B, C
F -> C
G -> D, E

Ancestors:
A ->
B -> A
C -> A
D -> B, A
E -> B, A, C
F -> C, A
G -> D, B, A, E, C

GOOD: B
BAD: E
ActualBAD: C

AFTER GOOD:
      C
     / \
D   E   F
 \ /
  G

THEN AFTER BAD:
      C
       \
        F

*/

// DAG implements the data structure of the DAG.
// The elements are literally just strings, because there is no reason for them not to be
// The parent relations are stored in "inboundEdge", with a map of the children to map of parents
type DAG struct {
	muDAG             sync.RWMutex
	vertices          map[string]bool
	inboundEdge       map[string]map[string]bool
	outboundEdge      map[string]map[string]bool
	numAncestorsCache map[string]int
	muCache           sync.RWMutex
	MostRecentBad     string
}

// ParamConfig is simply the configuration for the Midpoint selection
// Higher numbers = longer runtimes
// Your mission, should you choose to select it, is to modify these values to get as close as possible to the "ideal" score
type ParamConfig struct {
	// Limit is the limit, below which the very intensive "proper" midpoint selection will happen
	Limit int
	// Divisions is the number of samples to take in the "lightweight" midpoint selection
	Divisions int
}

// NewDAG creates / initializes a new DAG.
func NewDAG() *DAG {
	return &DAG{
		vertices:          make(map[string]bool),
		inboundEdge:       make(map[string]map[string]bool),
		outboundEdge:      make(map[string]map[string]bool),
		numAncestorsCache: make(map[string]int),
	}
}

// AddVertex adds the vertex v to the DAG. AddVertex returns an error, if v is
// nil, v is already part of the graph, or the id of v is already part of the
// graph.
func (d *DAG) AddVertex(v string) error {

	d.muDAG.Lock()
	defer d.muDAG.Unlock()

	// sanity checking
	if v == "" {
		return IdEmptyError{}
	}
	if _, exists := d.vertices[v]; exists {
		return IdDuplicateError{v}
	}
	d.addVertex(v)
	return nil
}

func (d *DAG) addVertex(v string) {
	d.vertices[v] = true
}

// VertexExists returns true if it exists, false otherwise
func (d *DAG) VertexExists(id string) (bool, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()

	if id == "" {
		return false, IdEmptyError{}
	}

	_, IDExists := d.vertices[id]
	if !IDExists {
		return false, IdUnknownError{id}
	}
	return true, nil
}

// DeleteVertex deletes the vertex v. DeleteVertex also deletes all attached
// edges (inbound and outbound) as well as ancestor- and descendant-caches of
// related vertices. DeleteVertex returns an error, if v is nil or unknown.
func (d *DAG) DeleteVertex(v string) error {

	d.muDAG.Lock()
	defer d.muDAG.Unlock()

	if err := d.saneVertex(v); err != nil {
		return err
	}

	// // delete v from inbound edges
	// for key := range d.inboundEdge {
	// 	delete(d.inboundEdge[key], v)
	// }

	// delete v in outbound edges of parents
	if _, exists := d.inboundEdge[v]; exists {
		for parent := range d.inboundEdge[v] {
			delete(d.outboundEdge[parent], v)
		}
	}

	// delete v in inbound edges of children
	if _, exists := d.outboundEdge[v]; exists {
		for child := range d.outboundEdge[v] {
			delete(d.inboundEdge[child], v)
		}
	}

	// delete in- and outbound of v itself
	delete(d.inboundEdge, v)
	delete(d.outboundEdge, v)

	// delete v itself
	delete(d.vertices, v)

	return nil
}

// AddEdge adds an edge between src and dst. AddEdge returns an error, if src
// or dst are nil or if the edge would create a loop. AddEdge calls AddVertex,
// if src and/or dst are not yet known within the DAG.
// src is the child, dst is the parent...
func (d *DAG) AddEdge(src string, dst string) error {

	d.muDAG.Lock()
	defer d.muDAG.Unlock()

	if src == "" || dst == "" {
		return IdEmptyError{}
	}
	if src == dst {
		return SrcDstEqualError{src, dst}
	}

	// ensure vertices
	if !d.vertices[src] {
		// TODO: I think that this is unnecessary
		if _, exists := d.vertices[src]; exists {
			return IdDuplicateError{src}
		}
		d.addVertex(src)
	}
	if !d.vertices[dst] {
		if _, exists := d.vertices[dst]; exists {
			return IdDuplicateError{dst}
		}
		d.addVertex(dst)
	}

	// if the edge is already known, there is nothing else to do
	if d.isEdge(src, dst) {
		return EdgeDuplicateError{src, dst}
	}

	// prepare d.outbound[src], iff needed
	if _, exists := d.outboundEdge[src]; !exists {
		d.outboundEdge[src] = make(map[string]bool)
	}

	// dst is a child of src
	d.outboundEdge[src][dst] = true

	// prepare d.inboundEdge[dst], iff needed
	if _, exists := d.inboundEdge[dst]; !exists {
		d.inboundEdge[dst] = make(map[string]bool)
	}

	// src is a parent of dst
	d.inboundEdge[dst][src] = true

	return nil
}

// IsEdge returns true, if there exists an edge between src and dst. IsEdge
// returns false if there is no such edge. IsEdge returns an error, if src or
// dst are nil, unknown or the same.
func (d *DAG) IsEdge(src string, dst string) (bool, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()

	if err := d.saneVertex(src); err != nil {
		return false, err
	}
	if err := d.saneVertex(dst); err != nil {
		return false, err
	}
	if src == dst {
		return false, SrcDstEqualError{src, dst}
	}

	return d.isEdge(src, dst), nil
}

func (d *DAG) isEdge(src string, dst string) bool {

	_, outboundExists := d.outboundEdge[src]
	_, inboundExists := d.inboundEdge[dst]

	return outboundExists && d.outboundEdge[src][dst] &&
		inboundExists && d.inboundEdge[dst][src]
}

// DeleteEdge deletes an edge. DeleteEdge also deletes ancestor- and
// descendant-caches of related vertices. DeleteEdge returns an error, if src
// or dst are nil or unknown, or if there is no edge between src and dst.
func (d *DAG) DeleteEdge(src string, dst string) error {

	d.muDAG.Lock()
	defer d.muDAG.Unlock()

	if err := d.saneVertex(src); err != nil {
		return err
	}
	if err := d.saneVertex(dst); err != nil {
		return err
	}
	if src == dst {
		return SrcDstEqualError{src, dst}
	}
	if !d.isEdge(src, dst) {
		return EdgeUnknownError{src, dst}
	}

	// delete inbound and outbound
	delete(d.inboundEdge[dst], src)
	delete(d.outboundEdge[src], dst)

	return nil
}

// GetOrder returns the number of vertices in the graph.
func (d *DAG) GetOrder() int {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	return len(d.vertices)
}

// GetSize returns the number of edges in the graph.
func (d *DAG) GetSize() int {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	return d.getSize()
}

func (d *DAG) getSize() int {
	count := 0
	for _, value := range d.outboundEdge {
		count += len(value)
	}
	return count
}

// GetRoots returns all vertices without parents.
func (d *DAG) GetRoots() map[string]bool {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	return d.getRoots()
}

func (d *DAG) getRoots() map[string]bool {
	roots := make(map[string]bool)
	for v := range d.vertices {
		srcIds, ok := d.inboundEdge[v]
		if !ok || len(srcIds) == 0 {
			roots[v] = true
		}
	}
	return roots
}

// GetLeafs returns all vertices without children.
func (d *DAG) GetLeafs() map[string]bool {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	return d.getLeafs()
}

func (d *DAG) getLeafs() map[string]bool {
	leafs := make(map[string]bool)
	for v := range d.vertices {
		dstIds, ok := d.outboundEdge[v]
		if !ok || len(dstIds) == 0 {
			leafs[v] = true
		}
	}
	return leafs
}

// GetVertices returns all vertices.
func (d *DAG) GetVertices() map[string]bool {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	return copyMap(d.vertices)
}

// GetParents returns all parents of vertex v. GetParents returns an error,
// if v is nil or unknown.
func (d *DAG) GetParents(v string) (map[string]bool, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	if err := d.saneVertex(v); err != nil {
		return nil, err
	}
	return copyMap(d.inboundEdge[v]), nil
}

// GetOrderedAncestors returns all ancestors of the vertex v in a breath-first
// order. Only the first occurrence of each vertex is returned.
// GetOrderedAncestors returns an error, if v is nil or unknown.
//
// Note, there is no order between sibling vertices. Two consecutive runs of
// GetOrderedAncestors may return different results.
func (d *DAG) GetOrderedAncestors(v string) ([]string, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	vertices, _, err := d.AncestorsWalker(v)
	if err != nil {
		return nil, err
	}
	var ancestors []string
	for v := range vertices {
		ancestors = append(ancestors, v)
	}
	return ancestors, nil
}

// AncestorsWalker returns a channel and subsequently returns / walks all
// ancestors of the vertex v in a breath first order. The second channel
// returned may be used to stop further walking. AncestorsWalker returns an
// error, if v is nil or unknown.
//
// Note, there is no order between sibling vertices. Two consecutive runs of
// AncestorsWalker may return different results.
func (d *DAG) AncestorsWalker(v string) (chan string, chan bool, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	if err := d.saneVertex(v); err != nil {
		return nil, nil, err
	}
	vertices := make(chan string)
	signal := make(chan bool, 1)
	go func() {
		d.muDAG.RLock()
		d.walkAncestors(v, vertices, signal)
		d.muDAG.RUnlock()
		close(vertices)
		close(signal)
	}()
	return vertices, signal, nil
}

func (d *DAG) walkAncestors(v string, vertices chan string, signal chan bool) {

	var fifo []string
	visited := make(map[string]bool)
	for parent := range d.inboundEdge[v] {
		visited[parent] = true
		fifo = append(fifo, parent)
	}
	for {
		if len(fifo) == 0 {
			return
		}
		top := fifo[0]
		fifo = fifo[1:]
		for parent := range d.inboundEdge[top] {
			if !visited[parent] {
				visited[parent] = true
				fifo = append(fifo, parent)
			}
		}
		select {
		case <-signal:
			return
		default:
			vertices <- top
		}
	}
}

// String return a textual representation of the graph.
func (d *DAG) String() string {
	result := fmt.Sprintf("DAG Vertices: %d - Edges: %d\n", d.GetOrder(), d.GetSize())
	result += fmt.Sprintf("Vertices:\n")
	d.muDAG.RLock()
	for k := range d.vertices {
		result += fmt.Sprintf("  %v\n", k)
	}
	result += fmt.Sprintf("Edges:\n")
	for v, parents := range d.inboundEdge {
		for parent := range parents {
			result += fmt.Sprintf("  %s -> %s\n", parent, v)
		}
	}
	d.muDAG.RUnlock()
	return result
}

func (d *DAG) saneVertex(v string) error {
	// sanity checking
	if v == "" {
		return IdEmptyError{}
	}
	_, exists := d.vertices[v]
	if !exists {
		return VertexUnknownError{v}
	}
	return nil
}

func copyMap(in map[string]bool) map[string]bool {
	out := make(map[string]bool)
	for key, value := range in {
		out[key] = value
	}
	return out
}

func copyBigMap(in map[string]map[string]bool) map[string]map[string]bool {
	out := make(map[string]map[string]bool)
	for key, value := range in {
		out[key] = value
	}
	return out
}

/***************************
********** Errors **********
****************************/

// VertexNilError is the error type to describe the situation, that a nil is
// given instead of a vertex.
type VertexNilError struct{}

// Implements the error interface.
func (e VertexNilError) Error() string {
	return fmt.Sprint("don't know what to do with 'nil'")
}

// IdEmptyError is the error type to describe the situation, that a nil is
// given instead of a vertex.
type IdEmptyError struct{}

// Implements the error interface.
func (e IdEmptyError) Error() string {
	return fmt.Sprint("don't know what to do with 'nil'")
}

// VertexDuplicateError is the error type to describe the situation, that a
// given vertex already exists in the graph.
type VertexDuplicateError struct {
	v string
}

// Implements the error interface.
func (e VertexDuplicateError) Error() string {
	return fmt.Sprintf("'%s' is already known", e.v)
}

// IdDuplicateError is the error type to describe the situation, that a given
// vertex id already exists in the graph.
type IdDuplicateError struct {
	v string
}

// Implements the error interface.
func (e IdDuplicateError) Error() string {
	return fmt.Sprintf("the id '%s' is already known", e.v)
}

// VertexUnknownError is the error type to describe the situation, that a given
// vertex does not exit in the graph.
type VertexUnknownError struct {
	v string
}

// Implements the error interface.
func (e VertexUnknownError) Error() string {
	return fmt.Sprintf("'%s' is unknown", e.v)
}

// IdUnknownError is the error type to describe the situation, that a given
// vertex does not exit in the graph.
type IdUnknownError struct {
	id string
}

// Implements the error interface.
func (e IdUnknownError) Error() string {
	return fmt.Sprintf("'%s' is unknown", e.id)
}

// EdgeDuplicateError is the error type to describe the situation, that an edge
// already exists in the graph.
type EdgeDuplicateError struct {
	src string
	dst string
}

// Implements the error interface.
func (e EdgeDuplicateError) Error() string {
	return fmt.Sprintf("edge between '%s' and '%s' is already known", e.src, e.dst)
}

// EdgeUnknownError is the error type to describe the situation, that a given
// edge does not exit in the graph.
type EdgeUnknownError struct {
	src string
	dst string
}

// Implements the error interface.
func (e EdgeUnknownError) Error() string {
	return fmt.Sprintf("edge between '%s' and '%s' is unknown", e.src, e.dst)
}

// EdgeLoopError is the error type to describe loop errors (i.e. errors that
// where raised to prevent establishing loops in the graph).
type EdgeLoopError struct {
	src string
	dst string
}

// Implements the error interface.
func (e EdgeLoopError) Error() string {
	return fmt.Sprintf("edge between '%s' and '%s' would create a loop", e.src, e.dst)
}

// SrcDstEqualError is the error type to describe the situation, that src and
// dst are equal.
type SrcDstEqualError struct {
	src string
	dst string
}

// Implements the error interface.
func (e SrcDstEqualError) Error() string {
	return fmt.Sprintf("src ('%s') and dst ('%s') equal", e.src, e.dst)
}

// ADDITIONAL STUFF

// GoodCommit should take the "good" commit, change the dag, and return an error if exists
// New dag should be the old dag - it and it's ancestors
func (d *DAG) GoodCommit(c string) error {
	// Get the ancestors
	ances, err := d.GetOrderedAncestors(c)
	if err != nil {
		return err
	}

	// Delete ancestors
	for _, result := range ances {
		err = d.DeleteVertex(result)
		if err != nil {
			return err
		}
	}
	// Delete itself
	err = d.DeleteVertex(c)
	if err != nil {
		return err
	}

	return nil
}

// BadCommit takes the "bad" commit, changes the dag, returning an error if required
// New dag should be it and it's ancestors
func (d *DAG) BadCommit(c string) error {
	d.MostRecentBad = c

	// Get the ancestors
	ances, err := d.GetOrderedAncestors(c)
	if err != nil {
		return err
	}

	// Get the vertices to remove
	verticesToRemove := d.GetVertices()
	for _, val := range ances {
		delete(verticesToRemove, val)
	}

	// Remove vertices we don't like any more
	for vertexToRemove := range verticesToRemove {
		err = d.DeleteVertex(vertexToRemove)
		if err != nil {
			return err
		}
	}

	return nil
}

// CommitAncestors is the
type CommitAncestors struct {
	Commit string
	Value  float64
}

// GetEstimateMidpointAgain gets a rough midpoint just based on the middle commit in the graph??
// DIVISIONS is the parameter to play around with, for the number of samples to take
func (d *DAG) GetEstimateMidpointAgain(c ParamConfig) (string, error) {

	leafs := d.GetLeafs()
	var maxValue CommitAncestors
	total := len(d.GetVertices())

	tovisit := make(map[string]bool)

	// log.Printf("Number of leaves: %v", len(leafs))

	// Go through all of the leafs (to cover all branches of the dag)
	for leaf := range leafs {
		// Get the ordered ancestors of this shit
		ancestors, err := d.GetOrderedAncestors(leaf)
		if err != nil {
			log.Print("Failed retrieving ancestors")
			return "", err
		}
		// log.Printf("Ancestors original length: %v", len(ancestors))

		// Get the middle half of the list
		start := (len(ancestors) / 5) * 2
		end := (len(ancestors) / 5) * 3
		ancestors = ancestors[start:end]

		// log.Printf("Start %v, end %v, length %v", start, end, len(ancestors))

		// Add these to a map
		INCREMENT := int(math.Max(1, float64(len(ancestors)/c.Divisions)))
		// log.Printf("INCREMENT: %v", INCREMENT)
		for i := 0; i < len(ancestors)-INCREMENT; i += INCREMENT {
			// log.Printf("attempt to access %v", i)
			tovisit[ancestors[i]] = true
		}
	}

	// Get the number of jobs
	numJobs := len(tovisit)

	// log.Printf("Number of jobs: %v", numJobs)

	// Spawn workers
	jobs := make(chan string, numJobs)
	results := make(chan CommitAncestors, numJobs)
	for w := 1; w <= runtime.GOMAXPROCS(0); w++ {
		go worker(w, d, jobs, results)
	}

	// log.Print("Workers spawned")

	// Submit jobs
	for k := range tovisit {
		jobs <- k
	}
	close(jobs)

	// log.Print("Jobs submitted")

	// Retrieve results
	for a := 1; a <= numJobs; a++ {
		result := <-results
		result.Value = math.Min(float64(result.Value), float64(total)-float64(result.Value))
		if result.Value >= maxValue.Value {
			maxValue = result
		}
	}
	close(results)

	return maxValue.Commit, nil
}

// GetMidPoint literally just returns the midpoint
func (d *DAG) GetMidPoint(c ParamConfig) (string, error) {

	if d.GetOrder() > c.Limit {
		log.Print("estimating...")
		return d.GetEstimateMidpointAgain(c)
	}

	var maxValue CommitAncestors

	temp := make(map[string]bool)

	temp = d.GetVertices()

	numJobs := len(temp)
	// log.Printf("Number of vertices to check: %v, but number of vertices? %v\n", numJobs, d.GetOrder())
	if numJobs == 1 {
		var thing string
		for s := range temp {
			thing = s
		}
		return thing, nil
	}

	jobs := make(chan string, numJobs)
	results := make(chan CommitAncestors, numJobs)

	// We want to spawn some workers, who have a recieving commit channel, and send back results to a channel.

	for w := 1; w <= runtime.GOMAXPROCS(0); w++ {
		go worker(w, d, jobs, results)
	}

	for j := range temp {
		jobs <- j
	}

	close(jobs)

	for a := 1; a <= numJobs; a++ {
		result := <-results
		result.Value = math.Min(float64(result.Value), float64(numJobs)-float64(result.Value))
		if result.Value >= maxValue.Value {
			maxValue = result
		}
	}
	close(results)

	// log.Printf("Midpoint is (%v)", maxValue.Commit)

	return maxValue.Commit, nil
}

func worker(id int, d *DAG, jobs <-chan string, results chan<- CommitAncestors) {
	for j := range jobs {
		ancs, err := d.GetOrderedAncestors(j)
		if err != nil {
			log.Fatal(err)
		}
		results <- CommitAncestors{
			Commit: j,
			Value:  float64(len(ancs)),
		}
	}
}
