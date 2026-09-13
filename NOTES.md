# goDB Notes

Where I went wrong. Any bugs I had and how I fixed them.
Documenting my understanding of databases and their components from zero to one.
Also my experience in go as I try to get a lot better at it.

### Pager

1. Learning how to program in go, I had the `seek()` function return values not be assigned to any variables therefore, I never knew it was running incorrectly. `_, err := file.seek(...)` is the correct usage, `file.seek(...)` was how I was using it. Any error I got just got lost and I never knew I was mixing up the parameters

2. What happens when a file crashes midwrite and we write a dirty page. I fixed it by having a copy/candidate of the page that we flush first and if any error arises we don't commit it to the file. Basically the lesson learned is to not to destroy any old data during an update.

3. I deliberately held off on implementing concurrency into the Pager, I want to build the B-tree module first before I try adding more complexity. Concurrency will be a later issue to worry about

4. Found a bug with how I was splitting nodes in a B-tree. I was aliasing the nodes slices of children and keys. Slice aliasing views back into the same slice. It doesn't create a copy if for whatever we write to the rightmost index of the left array it could overwrite or corrupt the right arrays keys and index

5. Found a bug where if we needed to add to the rightmost child on a non leaf node, it would error out since index would be > len(children). The issue was that I wasn't creating the left child nodes correctly which led to off by one error.
