Prudence: Extension Guide
=========================

### A note about versions

The Prudence `platform` package should maintain its contract between *minor* versions of
Prudence. I.e. extensions written against Prudence 3.1.6 should work with Prudence 3.1.12. The
latter may add more features, but should not remove or change the functionality of existing
ones. In other words, if a breaking change needs to be introduced to this package then the minor
version of Prudence would be bumped. Thus extensions written against Prudence 3.2.0 would not be
guaranteed to work with Prudence 3.1.x.

TODO

JavaScript APIs
---------------

JST Sugar
---------

Custom Objects
--------------

### Handlers

### Startables

Cache Backends
--------------
