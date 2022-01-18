package run

import "sort"

// PGroup collects actors (functions) and runs them concurrently.
// When one actor (function) returns, all actors are interrupted.
// The zero value of a Group is useful.
type PGroup struct {
	actors []actor
}

// Add an actor (function) to the group. Each actor must be pre-emptable by an
// interrupt function. That is, if interrupt is invoked, execute should return.
// Also, it must be safe to call interrupt even after execute has returned.
// Interrupt functions is invoked in the sort order.
func (g *PGroup) Add(execute func() error, interrupt func(error), interruptOrder int) {
	g.actors = append(g.actors, actor{
		execute:        execute,
		interrupt:      interrupt,
		interruptOrder: interruptOrder,
		index:          len(g.actors),
	})
}

// Run all actors (functions) concurrently.
// When the first actor returns, all others are interrupted.
// Run only returns when all actors have exited.
// Run returns the error returned by the first exiting actor.
func (g *PGroup) Run() error {
	if len(g.actors) == 0 {
		return nil
	}

	sort.Slice(g.actors, func(i, j int) bool {
		return g.actors[i].index < g.actors[j].index
	})

	errors := make(chan error, len(g.actors))
	for _, a := range g.actors {
		go func(a actor) {
			errors <- a.execute()
		}(a)
	}

	// Wait for the first actor to stop.
	err := <-errors

	sort.Slice(g.actors, func(i, j int) bool {
		return g.actors[i].interruptOrder < g.actors[j].interruptOrder
	})

	// Signal all actors to stop.
	for _, a := range g.actors {
		a.interrupt(err)
	}

	// Wait for all actors to stop.
	for i := 1; i < cap(errors); i++ {
		<-errors
	}

	// Return the original error.
	return err
}
