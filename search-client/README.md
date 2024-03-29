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

### Deploying the search client as a Wordpress plugin

0. Update version in ourroots.php and readme.txt

1. Run `npm run build` to build this project. This stores its output in a dist subdirectory of the project root.

2. Copy the files from dist into the plugin directory
    ```
    cp -r dist/img/* deploy/wp-plugin/ourroots/img
    cp dist/js/app*.js deploy/wp-plugin/ourroots/js/app.js
    cp dist/js/chunk-vendors*.js deploy/wp-plugin/ourroots/js/chunk-vendors.js
    cp dist/css/app*.css deploy/wp-plugin/ourroots/css/app.css
    cp dist/css/chunk-vendors*.css deploy/wp-plugin/ourroots/css/chunk-vendors.css
    ```
3. Either Create the plugin zip file and upload deploy/wp-plugin/ourroots.zip to your wordpress installation
    ```
    cd deploy/wp-plugin
    zip -r ourroots ourroots
    cd ../..
    ```
4. Or copy the files to the wordpress-plugin directory
   ```
   cp deploy/wp-plugin/ourroots/readme.txt ../../wordpress-plugin/trunk/
   cp deploy/wp-plugin/ourroots/ourroots.php ../../wordpress-plugin/trunk/
   cp deploy/wp-plugin/ourroots/includes/ourroots-class.php ../../wordpress-plugin/trunk/includes/
   cp deploy/wp-plugin/ourroots/css/* ../../wordpress-plugin/trunk/css/
   cp deploy/wp-plugin/ourroots/js/* ../../wordpress-plugin/trunk/js/
   cp -r deploy/wp-plugin/ourroots/img/* ../../wordpress-plugin/trunk/img/
   cp deploy/wp-plugin/ourroots/fonts/* ../../wordpress-plugin/trunk/fonts/
   cd ../../wordpress-plugin
   svn ci -m '<version>' --username <username> --password <password>
   ```


### Deploying the search client as a stand-alone application

1. Edit the .env.production file to point to where you deployed the cms server
2. Run `npm run build` to build this project
3. Copy the files in the dist directory to your hosting server. You could copy them to S3 for example: https://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteHosting.html
