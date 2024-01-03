# Spidey

![Tests Passing](https://github.com/anirudhsudhir/spidey/actions/workflows/test.yml/badge.svg)

A multithreaded web crawler written in Go.

Currently WIP.

### About:
Spidey has been built following the test-driven development approach.

### Results:
Spidey's best result has been a discovery of 26,454 unique links in 1 minute and 59 seconds, starting with seven seed URLs.  
All of these links were valid, including references to static content.  
It achieved this despite a predefined delay of 2 seconds between each consecutive set of network requests.  
This was enforced to avoid rate limit restrictions.  
