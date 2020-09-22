add_action('wp_enqueue_scripts','ourroots_search_load_scripts');
function ourroots_search_load_scripts() {
    if ( is_page_template('ourroots_search.php') ) {
        wp_enqueue_style('ourroots-search-fonts', get_stylesheet_directory_uri() . 'https://fonts.googleapis.com/css?family=Roboto:100,300,400,500,700,900', array(), '0.0.1');
        wp_enqueue_style('ourroots-search-icons', get_stylesheet_directory_uri() . 'https://cdn.jsdelivr.net/npm/@mdi/font@latest/css/materialdesignicons.min.css', array(), '0.0.1');
        wp_enqueue_style('ourroots-search-css-about', get_stylesheet_directory_uri() . '/ourroots_search/css/about.css', array(), '0.0.1');
        wp_enqueue_style('ourroots-search-css-chunk-vendors', get_stylesheet_directory_uri() . '/ourroots_search/css/chunk-vendors.css', array(), '0.0.1');
        wp_enqueue_style('ourroots-search-css-app', get_stylesheet_directory_uri() . '/ourroots_search/css/app.css', array(), '0.0.1');
    	wp_enqueue_script('ourroots-search-js-about', get_stylesheet_directory_uri() . '/ourroots_search/js/about.js', array(), '0.0.1');
	    wp_enqueue_script('ourroots-search-js-chunk-vendors', get_stylesheet_directory_uri() . '/ourroots_search/js/chunk-vendors.js', array(), '0.0.1', true);
        wp_enqueue_script('ourroots-search-js-app', get_stylesheet_directory_uri() . '/ourroots_search/js/app.js', array(), '0.0.1', true);
    }
}
