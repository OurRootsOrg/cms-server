<?php

/**
 * Plugin Name:       OurRootsDatabase
 * Description:       Integrate the OurRoots genealogical records database management system into WordPress.
 * Version:           1.0.25
 * Requires at least: 5.7
 * Requires PHP:      7.0
 * Author:            dallanq
 * Author URI:        https://www.linkedin.com/in/dallan-quass-7059/
 * License:           GPL v2 or later
 * License URI:       https://www.gnu.org/licenses/gpl-2.0.html
 * Text Domain:       jwto
 */

// Define plugin constants
define('OURROOTS_URL', plugin_dir_url(__FILE__));
define('OURROOTS_PATH', plugin_dir_path(__FILE__));
define('OURROOTS_ADMIN_SLUG', 'our-roots-settings');
define('OURROOTS_BASE', plugin_basename(__FILE__));
define('OURROOTS_OPTIONS_SLUG', 'ourroots_options');

require OURROOTS_PATH . 'includes/ourroots-class.php';