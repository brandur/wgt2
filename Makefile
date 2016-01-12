all: install scrape enrich render deploy

build:
	go build ./...

enrich: scrape
	wgt-enrich

deploy: render
# Note that AWS_ACCESS_KEY_ID will only be set for builds on the master
# branch because it's stored in `.travis.yml` as an encrypted variable.
# Encrypted variables are not made available to non-master branches because
# of the risk of being leaked through a script in a rogue pull request.
ifdef AWS_ACCESS_KEY_ID
	aws --version

	# Force text/html for HTML because we're not using an extension.
	aws s3 sync ./public/ s3://$(S3_BUCKET)/ --acl public-read --content-type text/html --delete --exclude 'assets*' $(AWS_CLI_FLAGS)

	# Then move on to assets and allow S3 to detect content type.
	aws s3 sync ./public/assets/ s3://$(S3_BUCKET)/assets/ --acl public-read --delete --follow-symlinks $(AWS_CLI_FLAGS)
endif

install:
	go install ./...

render: enrich
	wgt-render

scrape: install
	wgt-scrape

serve: install
	wgt-serve
