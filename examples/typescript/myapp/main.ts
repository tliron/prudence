
export function present() {
    context.response.contentType = 'text/html';
    this.write(`<html>
<body>
    Hello from TypeScript!
</body>
</html>
`);
}
