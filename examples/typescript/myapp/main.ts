
export function present(context: Context) {
    context.response.contentType = 'text/html';
    context.writeString(`<html>
<body>
    Hello from TypeScript!
</body>
</html>
`);
}
