# search client

## Project setup
```
npm install
```

### Compiles and hot-reloads for development
```
npm run serve
```

### Compiles and minifies for production
```
npm run build
```

### Lints and fixes files
```
npm run lint
```

### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).

## Deploying

### Deploying the search client to a custom Wordpress page template

1. Run `npm run build` to build this project.
2. Figure out which Wordpress theme you're using and navigate to the directory for that theme under wp-content/themes in your wordpress directory.
3. Add the code in deploy/ourroots_search_functions.php to the end of the functions.php file in the wordpress directory for your theme. 
   If you are upgrading a previous installation, you may want to change the version numbers from 0.0.1 to something else for cache-busting.
4. Copy deploy/ourroots_search.php to the wordpress directory for your theme.
5. Create a subdirectory ourroots_search under the wordpress directory for your theme.
6. Copy the dist/css, dist/js, and dist/img in this project to the new wordpress ourroots_search subdirectory.
7. Rename css/app...css to just app.css, and css/chunk-vendors...css to just css/chunk-vendors.css.
8. Rename js/app...js to just app.js, and js/chunk-vendors.js to just js/chunk-vendors.js.
9. Delete the js/...map files. 
10. In WordPress, create a new page (not a post, but a page) using ourroots_search as the page template. The page text can be empty.

### Deploying the search client as a stand-alone application

1. Edit the .env.production file to point to where you deployed the cms server
2. Run `npm run build` to build this project
3. Copy the files in the dist directory to your hosting server. You could copy them to S3 for example: https://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteHosting.html
