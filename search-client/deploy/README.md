# Deploying the search client to a custom Wordpress page template

1. Run `npm run build` to build this project.
2. Figure out which Wordpress theme you're using and navigate to the directory for that theme under wp-content/themes in your wordpress directory.
3. Add the code in ourroots_search_functions.php to the end of the functions.php file in the directory for your theme. 
   If you are upgrading a previous installation, you may want to change the version numbers from 0.0.1 to something else for cache-busting.
4. Copy ourroots_search.php to the directory for your theme.
5. Create a subdirectory ourroots_search under the directory for your theme.
6. Copy the css, js, and img subdirectories under the `dist` directory in this project to the ourroots_search subdirectory.
7. Rename css/about...css to just css/about.css, css/app...css to just app.css, and css/chunk-vendors...css to just css/chunk-vendors.css.
8. Rename js/about...js to just js/about.js, js/app...js to just app.js, and js/chunk-vendors.js to just js/chunk-vendors.js.
9. Delete the js/...map files. 
10. In WordPress, create a new page (not a post, but a page) using ourroots_search as the page template. The page text can be empty.
