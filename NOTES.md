1. Learning how to program in go, I had the `seek()` function return values not be assigned to any variables
   therefore, I never knew it was running incorrectly. `_, err := file.seek(...)` is the correct usage
   `file.seek(...)` was how I was using it. Any error I got just got lost and I never knew I was mixing up the parameters
2. What happens when a file crashes midwrite and we write a dirty page. I fixed it by having a copy/candidate of the page that we flush first and if any error arises we don't commit it to the file.
3. I deliberately held off on implementing concurrency into the Pager, I want to build the B-tree module first before I try adding more complexity. Concurrency will be a later isseu to worry about
