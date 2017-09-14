# Simple Static Blog in Golang

This project still being used for my blog. Well, the idea is to keep blog simple without database configuration and just serving what we wrote.

## What are being used
+ Golang
+ Markdown by russross https://github.com/russross/blackfriday
+ frontmatter by ericaro https://github.com/ericaro/frontmatter
+ httprouter by julienschmidt https://github.com/julienschmidt/httprouter
+ bpool by oxtoacart https://github.com/oxtoacart/bpool
+ bluemonday by microcosm-cc https://github.com/microcosm-cc/bluemonday

## What this blog does?
This blog loads the articles / posts into fake db and parsing it to the template when you load the page.
It have index (home), indexarticle which is links and titles to the article, aboutme that describing you more specific,
and other as you want it be.

## Pros
+ No database
+ Just place your markdown file to ***posts*** folder and restart the server, its done to update your article.
+ Auto indexing all article by date

## Cons
+ Need to restart everytime you update post
+ Still learn to make it clean code
+ Not tested yet
