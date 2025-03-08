build:
	@go build -o ./dist/server .

run: build
	@./dist/server

generate-templates:
	@templ generate

generate-styles:
	npx tailwindcss -i ./views/input.css -o ./static/vendor/tailwind.css --minify

# re-create _templ.txt files on change, then send reload event to browser. 
# Default url: http://localhost:7331
live/templ:
	templ generate --watch --proxy="http://localhost:1323" --open-browser=false -v

# run air to detect any go file changes to re-build and re-run the server.
live/server:
	air

# run tailwindcss to generate the styles.css bundle in watch mode.
live/tailwind:
	air \
	--build.cmd "npx @tailwindcss/cli -i ./views/input.css -o ./static/vendor/tailwind.css --minify && templ generate --notify-proxy" \
	--build.bin "true" \
	--build.delay "100" \
	--build.exclude_dir "" \
	--build.include_dir "views" \
	--build.include_ext "templ"


# watch for any js or css change in the assets/ folder, then reload the browser via templ proxy.
live/sync_assets:
	sleep 1 && air \
	--build.cmd "templ generate --notify-proxy" \
	--build.bin "true" \
	--build.delay "500" \
	--build.exclude_dir "" \
	--build.include_dir "public" \
	--build.include_ext "js,css"

dev: 
	make -j4 live/templ live/server live/tailwind live/sync_assets
