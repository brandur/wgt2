# wgt2

[![Travis status](https://travis-ci.org/brandur/wgt2.svg?branch=master)](https://travis-ci.org/brandur/wgt2)

## Install

    make install

## Obtain Refresh Token

In `.env`:

    CLIENT_ID=...
    CLIENT_SECRET=...

Then:

    forego run wgt-procure

And add the resulting refresh token to `.env`:

    REFRESH_TOKEN=...

### Generate

Scrape artist list from the WGT site, enrichen it using the Spotify API, and
then build static artifacts:

    wgt-scrape
    forego run wgt-enrich
