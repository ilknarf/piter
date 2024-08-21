# PIter

## Parallel Consumers using Go Iterators

- Required Go version: `1.23+`

### Explanation
As seen [here](https://go.dev/blog/range-functions#passing-a-function-to-a-push-iterator), Go's new push iterators can take in arbitrary `yield` functions. Taking advantage of this, we can easily create iterator consumers that
:
1. Have simple Go error handling semantics.
2. Can process jobs in parallel using a set number of workers.
3. Utilize existing idiomatic and time-tested patterns to do so (`errgroup.Group`).

### Use case
- This library's goal is to provide a simple surface to enable parallel iterator consumption. Potential use cases:
    - Batch processing
	- Consuming streams in a worker pool
	- Lambda handlers (Is this even optimal?)
- Perfomance implications aside, this library can also enable incremental refactoring of code through simplification of iterative code into declarative processing.

### Caveats
- `slices.PFlatMap` and`slices.PMap` DO NOT preserve order.
- `maps.PMap` overwrites duplicate keys in an undeterministic fashion. This is irrelevant when processing maps or slices, but more important when processing `iter.Seq2` with different semantics.

### Usage
```go
package main

import (
	"slices"

	pslices "github.com/ilknarf/piter/slices"
)

func taskHandler(ctx context.Context, ids []string) (*Result, error) {
	// do important stuff
	return res, nil
}

func main() {
	ctx := context.Background()
	ids := []string{"id1", "id2", "id3", "id4", "id5"} // ...
	batchIter := slices.Chunk(ids)

	// number of threads. this is directly passed to `errgroup.Group.SetLimit`
	batchSize := 10

	// in parallel
	res, err := pslices.PMap(ctx, batchIter, batchSize, taskHandler)

	// ...
}

```

See [examples](examples/) for more... examples.
