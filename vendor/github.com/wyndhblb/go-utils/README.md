# goutils


Ideally you'd clone the usual ...


    go get github.com/wyndhblb/goutils
   

Due to some vendor dependencies, then simply import what you wish.

All of these things are really too simple to matter, but they are here so i don't constantly have to repeat myself.


## bitstream

From https://github.com/dgryski/go-tsz/blob/master/bstream.go but "public" so that it can be used in many a place for many things.

Hats off dgryski.

This is the least "simple" thing here (at least for those that can't bit shift in their heads).

## once

Ever not want to start something (or stop) but make sure start did not happen more then once? Say hello to StartStop.

## pools

If you're using sync.Pool, I bet you've created a zillion of these (byte, buffer, mutexes, wait groups) .. put them in a spot.

## shutdown

Having many go routines to deal with their shutdown in a nice way from a SIGINT can be painful. Especially if the routines
are very disjointed.  Not really connected to other ones via some channel mechanism as that would require a very large 
select/case tree of large proportion, when all you want is to get a "Poison Pill message" and let things empty their queues 
and finish. This is basically a global wait group for that action.

## tomlenv

A missing part of TOML, simply pull some singular vars from the ENV (aka docker for like things) inside a very large 
complicated config file.

## options

A really simple map like entity for config options, that will check value types when getting them.

## shared

Yep a global "super map of stuff".  I know it's "frowned" upon, but just like shutdown, go is a compiled language, native,
ASM lang at its core.  This can be really useful for performance, and pushing/read a nodes config options around the 
farms (this is not zookeeper, it's internal, but lets one set a submodule's state for easy reading by something else).
 
## sorts

The fact that `sort.Sort` cannot natively sort the basic types is, well, ..., so the basic u/int64->8 have sorts.