# Spidey

![Tests Passing](https://github.com/anirudhsudhir/spidey/actions/workflows/test.yml/badge.svg)

A multithreaded web crawler written in Go.

NOTE: An improved version of the same project can be found at [Spidey-v2](https://github.com/anirudhsudhir/Spidey-v2)

### About

Spidey has been built following the test-driven development approach.

### Usage

1. Clone this repository

```bash
git clone https://github.com/anirudhsudhir/spidey.git
cd spidey
```

2. Create a "seeds.txt" and add the seed links in quotes consecutively

    Sample seeds.txt

```text
"http://example.com"
"https://abcd.com"
```

3. Build the project and run Spidey.
   Pass the total allowed runtime time of the crawler and maximum request time per link (in milliseconds) as arguments

```bash
go build
./spidey 5000 2000
```

### Results

Spidey's best result has been a discovery of 26,454 unique links in 1 minute and 59 seconds, starting with seven seed URLs.  
All of these links were valid, including references to static content.  
It achieved this despite a predefined delay of 2 seconds between each consecutive set of network requests.  
This was enforced to avoid rate limit restrictions.
