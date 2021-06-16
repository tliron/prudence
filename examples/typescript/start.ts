
import {router} from './myapp/router';

declare var prudence: any;

prudence.start(new prudence.Server({
    address: 'localhost:8080',
    handler: router
}));
