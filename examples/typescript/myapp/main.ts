
declare var prudence: any;

export function present(context: any) {
    context.response.contentType = 'text/html';
    context.writeString(`<html>
<body>
    Hello from TypeScript!
</body>
</html>
`);
}
