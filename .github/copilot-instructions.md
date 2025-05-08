Always generate ideomatic golang code.
Write comments as full sentences, starting with the name of the entity being described, to enhance clarity and documentation quality.
Pass `context.Context` as the first parameter to functions that require it, and avoid embedding `Context` within structs.
Prefer declaring nil slices using `var s []T` over empty slices like `s := []T{}` to adhere to idiomatic Go practices.
Utilize `crypto/rand` for generating cryptographic keys or tokens, avoiding `math/rand` for security-sensitive operations.
Employ Go's built-in `testing` package for writing unit tests to ensure compatibility and leverage Go's testing tools effectively.
Favor interfaces to achieve polymorphism instead of relying on reflection, which can be less efficient and harder to maintain.
Compose small, focused types to build complex functionality, aligning with Go's design philosophy of composition over inheritance.
Use goroutines judiciously, ensuring proper synchronization using channels or `sync.WaitGroup` to avoid race conditions and resource leaks.
Organize the project by separating executable commands and library code into distinct packages to promote modularity and reusability.
Write clear, descriptive comments outlining the functionality.
Always generate tests with standard library.
Do not use github.com/stretchr/testify/assert for assertions.