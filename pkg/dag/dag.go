// Package dag implements directed acyclic graphs (DAGs).
package dag

import (
	"fmt"
	"sync"
)

// // Vertex is the interface to be implemented for the vertices of the DAG.
// type Vertex interface {
// 	// Return the id of this vertex. This id must be unique and never change.
// 	String() string
// }

// DAG implements the data structure of the DAG.
// The elements are literally just strings, because there is no reason for them not to be
// The parent relations are stored in "inboundEdge", with a map of the children to map of parents
type DAG struct {
	muDAG          sync.RWMutex
	vertices       map[string]bool
	inboundEdge    map[string]map[string]bool
	muCache        sync.RWMutex
	verticesLocked *dMutex
	ancestorsCache map[string]map[string]bool
	visited        map[string]bool
}

// NewDAG creates / initializes a new DAG.
func NewDAG() *DAG {
	return &DAG{
		vertices:       make(map[string]bool),
		inboundEdge:    make(map[string]map[string]bool),
		verticesLocked: newDMutex(),
		ancestorsCache: make(map[string]map[string]bool),
		visited:        make(map[string]bool),
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

	// // get descendents and ancestors as they are now
	// ancestors := copyMap(d.getAncestors(v))

	// // delete v in outbound edges of parents
	// if _, exists := d.inboundEdge[v]; exists {
	// 	for parent := range d.inboundEdge[v] {
	// 		delete(d.outboundEdge[parent], v)
	// 	}
	// }

	// // delete v in inbound edges of children
	// if _, exists := d.outboundEdge[v]; exists {
	// 	for child := range d.outboundEdge[v] {
	// 		delete(d.inboundEdge[child], v)
	// 	}
	// }

	// delete v from inbound edges
	for key := range d.inboundEdge {
		delete(d.inboundEdge[key], v)
	}

	// delete in- and outbound of v itself
	delete(d.inboundEdge, v)

	// // for v and all its descendants delete cached ancestors
	// for descendant := range descendants {
	// 	if _, exists := d.ancestorsCache[descendant]; exists {
	// 		delete(d.ancestorsCache, descendant)
	// 	}
	// }
	delete(d.ancestorsCache, v)

	// // for v and all its ancestors delete cached descendants
	// for ancestor := range ancestors {
	// 	if _, exists := d.descendantsCache[ancestor]; exists {
	// 		delete(d.descendantsCache, ancestor)
	// 	}
	// }
	// delete(d.descendantsCache, v)

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
		if exists, _ := d.VertexExists(src); exists {
			return IdDuplicateError{src}
		}
		d.addVertex(src)
	}
	if !d.vertices[dst] {
		if exists, _ := d.VertexExists(dst); exists {
			return IdDuplicateError{dst}
		}
		d.addVertex(dst)
	}

	// if the edge is already known, there is nothing else to do
	if d.isEdge(src, dst) {
		return EdgeDuplicateError{src, dst}
	}

	// get ancestors as they are now
	// ancestors := copyMap(d.getAncestors(src))

	// // prepare d.outbound[src], iff needed
	// if _, exists := d.outboundEdge[src]; !exists {
	// 	d.outboundEdge[src] = make(map[Vertex]bool)
	// }

	// // dst is a child of src
	// d.outboundEdge[src][dst] = true

	// prepare d.inboundEdge[dst], iff needed
	if _, exists := d.inboundEdge[dst]; !exists {
		d.inboundEdge[dst] = make(map[string]bool)
	}

	// src is a parent of dst
	d.inboundEdge[dst][src] = true

	// // for dst and all its descendants delete cached ancestors
	// for descendant := range descendants {
	// 	if _, exists := d.ancestorsCache[descendant]; exists {
	// 		delete(d.ancestorsCache, descendant)
	// 	}
	// }
	delete(d.ancestorsCache, dst)

	// // for src and all its ancestors delete cached descendants
	// for ancestor := range ancestors {
	// 	if _, exists := d.descendantsCache[ancestor]; exists {
	// 		delete(d.descendantsCache, ancestor)
	// 	}
	// }
	// delete(d.descendantsCache, src)

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

	// get ancestors as they are now
	// ancestors := copyMap(d.getAncestors(dst))

	// delete inbound
	delete(d.inboundEdge[dst], src)

	// // for src and all its descendants delete cached ancestors
	// for descendant := range descendants {
	// 	if _, exists := d.ancestorsCache[descendant]; exists {
	// 		delete(d.ancestorsCache, descendant)
	// 	}
	// }
	delete(d.ancestorsCache, src)

	// // for dst and all its ancestors delete cached descendants
	// for ancestor := range ancestors {
	// 	if _, exists := d.descendantsCache[ancestor]; exists {
	// 		delete(d.descendantsCache, ancestor)
	// 	}
	// }
	// delete(d.descendantsCache, dst)

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

// GetAncestors return all ancestors of the vertex v. GetAncestors returns an
// error, if v is nil or unknown.
//
// Note, in order to get the ancestors, GetAncestors populates the ancestor-
// cache as needed. Depending on order and size of the sub-graph of v this may
// take a long time and consume a lot of memory.
func (d *DAG) GetAncestors(v string) (map[string]bool, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	if err := d.saneVertex(v); err != nil {
		return nil, err
	}
	return copyMap(d.getAncestors(v)), nil
}

func (d *DAG) getAncestors(v string) map[string]bool {

	// in the best case we have already a populated cache
	d.muCache.RLock()
	cache, exists := d.ancestorsCache[v]
	d.muCache.RUnlock()
	if exists {
		return cache
	}

	// lock this vertex to work on it exclusively
	d.verticesLocked.lock(v)
	defer d.verticesLocked.unlock(v)

	// now as we have locked this vertex, check (again) that no one has
	// meanwhile populated the cache
	d.muCache.RLock()
	cache, exists = d.ancestorsCache[v]
	d.muCache.RUnlock()
	if exists {
		return cache
	}

	// as there is no cache, we start from scratch and first of all collect
	// all ancestors locally
	cache = make(map[string]bool)
	var mu sync.Mutex
	if parents, ok := d.inboundEdge[v]; ok {

		// for each parent collect its ancestors
		for parent := range parents {
			parentAncestors := d.getAncestors(parent)
			mu.Lock()
			for ancestor := range parentAncestors {
				cache[ancestor] = true
			}
			cache[parent] = true
			mu.Unlock()
		}
	}

	// remember the collected descendents
	d.muCache.Lock()
	d.ancestorsCache[v] = cache
	d.muCache.Unlock()
	return cache
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

// // ReduceTransitively transitively reduce the graph.
// //
// // Note, in order to do the reduction the descendant-cache of all vertices is
// // populated (i.e. the transitive closure). Depending on order and size of DAG
// // this may take a long time and consume a lot of memory.
// func (d *DAG) ReduceTransitively() {

// 	d.muDAG.Lock()
// 	defer d.muDAG.Unlock()

// 	graphChanged := false

// 	// populate the descendents cache for all roots (i.e. the whole graph)
// 	for root := range d.getRoots() {
// 		_ = d.getDescendants(root)
// 	}

// 	// for each vertex
// 	for v := range d.vertices {

// 		// map of descendants of the children of v
// 		descendentsOfChildrenOfV := make(map[Vertex]bool)

// 		// for each child of v
// 		for childOfV := range d.outboundEdge[v] {

// 			// collect child descendants
// 			for descendent := range d.descendantsCache[childOfV] {
// 				descendentsOfChildrenOfV[descendent] = true
// 			}
// 		}

// 		// for each child of v
// 		for childOfV := range d.outboundEdge[v] {

// 			// remove the edge between v and child, iff child is a
// 			// descendants of any of the children of v
// 			if descendentsOfChildrenOfV[childOfV] {
// 				delete(d.outboundEdge[v], childOfV)
// 				delete(d.inboundEdge[childOfV], v)
// 				graphChanged = true
// 			}
// 		}
// 	}

// 	// flush the descendants- and ancestor cache if the graph has changed
// 	if graphChanged {
// 		d.flushCaches()
// 	}
// }

// FlushCaches completely flushes the descendants- and ancestor cache.
//
// Note, the only reason to call this method is to free up memory.
// Otherwise the caches are automatically maintained.
func (d *DAG) FlushCaches() {
	d.muDAG.Lock()
	defer d.muDAG.Unlock()
	d.flushCaches()
}

func (d *DAG) flushCaches() {
	d.ancestorsCache = make(map[string]map[string]bool)
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
