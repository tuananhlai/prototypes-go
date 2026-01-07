What is the best way to implement PATCH endpoint in Go?
1. Use a struct with nullable fields (pointer, Null*, ...) -> Generate and run UPDATE queries based on non-null fields.
2. Accept a JSON PATCH object -> retrieve the current object -> Marshal it to JSON -> apply JSON patch -> unmarshal the modified json back into a struct -> write to db.
3. Accept PUT request instead.
