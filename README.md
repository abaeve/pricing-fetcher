# pricing-fetcher

[![Coverage Status](https://coveralls.io/repos/github/abaeve/pricing-fetcher/badge.svg?branch=master)](https://coveralls.io/github/abaeve/pricing-fetcher?branch=master)
[![Build Status](https://drone.aba-eve.com/api/badges/abaeve/pricing-fetcher/status.svg)](https://drone.aba-eve.com/abaeve/pricing-fetcher)

Backend worker that will periodically poll CCP Datasources for market prices and publish internally.

Still very much a work in progress but the idea is fairly complex.

Some constraints I'm putting on it:

1. It must be threaded
    1. Must be able to download multiple pages concurrently
2. It must be poolable
    1. Meaning it must have a way for _something_ to limit how much it can suck on the internets
3. Must be asynchronous
    1. The transactions it's downloading are not sticking around in memory for very long
4. The state of the regions download must be known and publishable
    1. Meaning it can send a message when it starts and has completely finished