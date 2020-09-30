# Deploying the search client to a custom Wordpress page template

1. Run `npm run build` to build this project. This stores its output in a dist subdirectory of the project root.
2. Use Filezilla or another ftp client to connect to your wordpress server.
3. Figure out which wordpress theme you're using and navigate to the directory for that theme under wp-content/themes in your wordpress directory.
4. Add the code in ourroots_search_functions.php from this deploy directory to the end of the functions.php file in the directory for your theme.
   If you are upgrading a previous installation, you may want to change the version numbers from 0.0.1 to something else for cache-busting.
5. Copy ourroots_search.php from this deploy directory to the wordpress directory for your theme.
5. Create a subdirectory ourroots_search under the wordpress directory for your theme.
6. Copy the contents of the dist directory in this project to ourroots_search. 
   You should end up with three new subdirectories under ourroots_search: css, img, and js
7. Rename ourroots_search/css/app...css to just app.css, and ourroots_search/css/chunk-vendors...css to just chunk-vendors.css
8. Rename ourroots_search/js/app...js to just app.js, and ourroots_search/js/chunk-vendors...js to just chunk-vendors.js
9. In wordpress, create a new page (not a post, but a page) using ourroots_search as the page template. The page text can be empty.
