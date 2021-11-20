
import {resource} from './myapp/resource';

prudence.start(new prudence.Server({
    address: ':8080',
    handler: resource
}));
