# wgt2

[![Travis status](https://travis-ci.org/brandur/wgt2.svg?branch=master)](https://travis-ci.org/brandur/wgt2)

An app that scrapes the [WGT][wgt] website and compiles artist information for
easier digestion.

## Install

    make install

## Obtain Refresh Token

    export CLIENT_ID=...
    export CLIENT_SECRET=...

Then authorize the app under your account and get a set of tokens issued with:

    wgt-procure

Then make sure to export your refresh token:

    export REFRESH_TOKEN=...

## Generate

Scrape artist list from the WGT site, enrichen it using the Spotify API, and
then build static artifacts:

    wgt-scrape
    wgt-enrich
    wgt-render

[wgt]: http://www.wave-gotik-treffen.de/english/bands.php
