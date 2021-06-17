
import {resource} from './myapp/resource';

declare const prudence: any;

prudence.start(new prudence.Server({
    address: 'localhost:8080',
    handler: resource
}));
