# rwlockversuschannels

*This repository is for learning purposes only.*

Prometheus uses read/write locks for controlling access to writes and reads on its memory series. Also the memory structure for these locks adds a padding so that it takes up to 64 bytes per lock. This is to avoid having more than one lock in the same cache line.

I'm concerned that in very big instances with hundreds of thousands or millions of active series there could be too many locks in place leading into cache contention and reader starvation.

I don't have data that validates this hypothesis there is so much to learn about how these things work underneath that I figured it would be best to start playing with a smaller scenario.
