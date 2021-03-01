<?php
    /* Settings page options */

    require_once( OURROOTS_PATH . 'hd-wp-settings-api/class-hd-wp-settings-api.php' ); // Settings API

    $jwto_options = array(
        'page_title'  => __( 'OurRootsCMS', 'zipf' ),
        'menu_title'  => __( 'OurRootsCMS', 'zipf' ),
        'menu_slug'   => OURROOTS_OPTIONS_SLUG,
        'capability'  => 'manage_options',
        'icon'        => 'dashicons-admin-generic',
        'position'    => 61
    );

    $jwto_fields = array(
        'hd_tab_1'      => array(
            'title' => __( 'Settings', 'zipf' ),
            'type'  => 'tab',
        ),
        'jwto_enable' => array(
            'title'   => __( 'Enable shortcode?', 'zipf' ),
            'type'    => 'checkbox',
            'default' => 0,
            'desc'    => __( 'Enable shortcode?', 'zipf' ),
            'sanit'   => 'nohtml',
        ),
        'jwto_society_id' => array(
            'title'   => __( 'Society ID', 'zipf' ),
            'type'    => 'number',
            'default' => '',
            'min' => 1,
            'max' => 9999,
            'desc'    => __( 'Society ID', 'zipf' ),
            'sanit'   => 'nohtml',
        ),
        'jwto_secret' => array(
            'title'   => __( 'Secret key', 'zipf' ),
            'type'    => 'text',
            'default' => '',
            'desc'    => __( 'Secret key', 'zipf' ),
            'sanit'   => 'nohtml',
        ),
        'jwto_token_expire' => array(
            'title'   => __( 'Token expiration days', 'zipf' ),
            'type'    => 'number',
            'min'    => 1,
            'max'    => 14,
            'default' => '',
            'desc'    => __( 'Token expiration days', 'zipf' ),
            'sanit'   => 'nohtml',
        ),
        'jwto_custom_css' => array(
            'title'   => __( 'Custom styles', 'zipf' ),
            'type'    => 'textarea',
            'default' => '',
            'desc'    => __( 'Custom styles', 'zipf' ),
        ),

    );

    $jwto_settings = new OURROOTS_WP_Settings_API( $jwto_options, $jwto_fields );