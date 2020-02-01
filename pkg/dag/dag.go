// Package dag implements directed acyclic graphs (DAGs).
package dag

import (
	"fmt"
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

*/

// DAG implements the data structure of the DAG.
// The elements are literally just strings, because there is no reason for them not to be
// The parent relations are stored in "inboundEdge", with a map of the children to map of parents
type DAG struct {
	muDAG          sync.RWMutex
	vertices       map[string]bool
	inboundEdge    map[string]map[string]bool
	outboundEdge   map[string]map[string]bool
	muCache        sync.RWMutex
	verticesLocked *dMutex
	// ancestorsCache map[string]map[string]bool
	visited       map[string]bool
	visitedLock   sync.RWMutex
	MidPoint      string
	MostRecentBad string
}

// NewDAG creates / initializes a new DAG.
func NewDAG() *DAG {
	return &DAG{
		vertices:       make(map[string]bool),
		inboundEdge:    make(map[string]map[string]bool),
		outboundEdge:   make(map[string]map[string]bool),
		verticesLocked: newDMutex(),
		// ancestorsCache: make(map[string]map[string]bool),
		visited: make(map[string]bool),
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

	_, inboundExists := d.inboundEdge[dst]

	return inboundExists && d.inboundEdge[dst][src]
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
	for _, value := range d.inboundEdge {
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

// GetAncestorsSimple returns all ancestors of the vertex v, or an error if
// v is nil or unknown
func (d *DAG) GetAncestorsSimple(v string) (map[string]bool, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	if err := d.saneVertex(v); err != nil {
		return nil, err
	}

	// Reset Visited
	d.visited = make(map[string]bool)

	d.getAncestorsSimple(&v)

	return d.visited, nil
}

func (d *DAG) getAncestorsSimple(v *string) bool {
	d.visitedLock.Lock()
	if _, visited := d.visited[*v]; visited {
		d.visitedLock.Unlock()
		return false
	}
	d.visited[*v] = true
	d.visitedLock.Unlock()

	if parents, ok := d.inboundEdge[*v]; ok {

		// for each parent collect its ancestors
		for parent := range parents {
			// I'll be honest, I don't care about the output - the returns just kill the function quickly
			d.getAncestorsSimple(&parent)
		}
		return true
	}
	return false
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

/***************************
********** dMutex **********
****************************/

type cMutex struct {
	mutex sync.Mutex
	count int
}

// Structure for dynamic mutexes.
type dMutex struct {
	mutexes     map[interface{}]*cMutex
	globalMutex sync.Mutex
}

// Initialize a new dynamic mutex structure.
func newDMutex() *dMutex {
	return &dMutex{
		mutexes: make(map[interface{}]*cMutex),
	}
}

// Get a lock for instance i
func (d *dMutex) lock(i interface{}) {

	// acquire global lock
	d.globalMutex.Lock()

	// if there is no cMutex for i, create it
	if _, ok := d.mutexes[i]; !ok {
		d.mutexes[i] = new(cMutex)
	}

	// increase the count in order to show, that we are interested in this
	// instance mutex (thus now one deletes it)
	d.mutexes[i].count++

	// remember the mutex for later
	mutex := &d.mutexes[i].mutex

	// as the cMutex is there, we have increased the count and we know the
	// instance mutex, we can release the global lock
	d.globalMutex.Unlock()

	// and wait on the instance mutex
	(*mutex).Lock()
}

// Release the lock for instance i.
func (d *dMutex) unlock(i interface{}) {

	// acquire global lock
	d.globalMutex.Lock()

	// unlock instance mutex
	d.mutexes[i].mutex.Unlock()

	// decrease the count, as we are no longer interested in this instance
	// mutex
	d.mutexes[i].count--

	// if we where the last one interested in this instance mutex delete the
	// cMutex
	if d.mutexes[i].count == 0 {
		delete(d.mutexes, i)
	}

	// release the global lock
	d.globalMutex.Unlock()
}

// ADDITIONAL STUFF

// GoodCommit
// BadCommit
// MidPoint
// MostRecentBadCommit

// GoodCommit should take the "good" commit, change the dag, and return an error if exists
// New dag should be the old dag - it and it's ancestors
// Should also return the next commit to test, or the answer if:
// There are no edges left, it returns the last known bad commit
// There is one vertex left, it returns that
func (d *DAG) GoodCommit(c string) error {
	// Get the ancestors
	ances, err := d.GetOrderedAncestors(c)
	if err != nil {
		return err
	}

	// d.MidPoint = ances[len(ances)/2]
	for _, result := range ances {
		err = d.DeleteVertex(result)
		if err != nil {
			return err
		}
	}
	err = d.DeleteVertex(c)
	if err != nil {
		return err
	}

	// WORK OUT A MIDPOINT

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

	d.MidPoint = ances[len(ances)/2]

	ances = append(ances, c)
	for key := range d.inboundEdge {
		if !contains(ances, key) {
			err = d.DeleteVertex(key)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
