
function represent(context) {
    context.log.info('age!');
    context.contentType = 'text/html';
    var md = prudence.load('age.md');
    md = prudence.render(md, 'markdown');
    context.write(md);
}
