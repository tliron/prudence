
declare var prudence: any;

export const router = new prudence.Router({
    routes: [{
        handler: function(context) {
            context.writeString('Hello from TypeScript\n');
        }
    }]
});
