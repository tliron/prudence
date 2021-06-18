
import {resource} from './myapp/resource';

prudence.start(new prudence.Server({
    address: 'localhost:8080',
    handler: resource
}));
