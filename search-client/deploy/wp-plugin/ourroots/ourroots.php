<?php

/**
 * Plugin Name: OurRootsDatabase
 * Description: Databases for genealogy societies
 * Author: OurRoots.org
 * Version: 1.0.5
 * Text Domain: jwto
 */

// Define plugin constants
define('OURROOTS_URL', plugin_dir_url(__FILE__));
define('OURROOTS_PATH', plugin_dir_path(__FILE__));
define('OURROOTS_ADMIN_SLUG', 'helpc-settings');
define('OURROOTS_BASE', plugin_basename(__FILE__));
define('OURROOTS_OPTIONS_SLUG', 'ourroots_options');

require OURROOTS_PATH . 'ourroots-settings.php';
require OURROOTS_PATH . 'includes/ourroots-class.php';