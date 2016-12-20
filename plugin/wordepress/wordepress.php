<?php
/*
Plugin Name: Weaveworks Wordepress
Description: Host technical documentation in WordPress
Version: 1.1.0
Author: Adam Harrison
*/

function wordepress_init_post_type( $post_type_name ) {
    global $wp_post_types;

    if( isset( $wp_post_types[ $post_type_name ] ) ) {
        $wp_post_types[$post_type_name]->show_in_rest = true;
        $wp_post_types[$post_type_name]->rest_base = $post_type_name;
        $wp_post_types[$post_type_name]->rest_controller_class = 'WP_REST_Posts_Controller';
    }
}

// Register the `document` post type with the REST API
add_action( 'init', function () { wordepress_init_post_type ( 'documentation' ); }, 100);

// Register the `tutorial` post type with the REST API
add_action( 'init', function() { wordepress_init_post_type ( 'tutorials' ); }, 100);

add_action( 'rest_api_init', function () {
    register_rest_field( 'documentation',
        'wpcf-product',
        array(
            'get_callback'    => 'wordepress_get_meta',
            'update_callback' => 'wordepress_update_meta',
            'schema'          => null,
        )
    );
    register_rest_field( 'documentation',
        'wpcf-version',
        array(
            'get_callback'    => 'wordepress_get_meta',
            'update_callback' => 'wordepress_update_meta',
            'schema'          => null,
        )
    );
    register_rest_field( 'documentation',
        'wpcf-name',
        array(
            'get_callback'    => 'wordepress_get_meta',
            'update_callback' => 'wordepress_update_meta',
            'schema'          => null,
        )
    );
    register_rest_field( 'documentation',
        'wpcf-tag',
        array(
            'get_callback'    => 'wordepress_get_meta',
            'update_callback' => 'wordepress_update_meta',
            'schema'          => null,
        )
    );
});

add_action( 'rest_api_init', function () {
    register_rest_field( 'tutorials',
        'wpcf-name',
        array(
            'get_callback'    => 'wordepress_get_meta',
            'update_callback' => 'wordepress_update_meta',
            'schema'          => null,
        )
    );
});

function wordepress_get_meta( $object, $field_name, $request ) {
    return get_post_meta( $object[ 'id' ], $field_name, true );
}

function wordepress_update_meta( $value, $object, $field_name ) {
    if ( ! $value || ! is_string( $value ) ) {
        return;
    }

    return update_post_meta( $object->ID, $field_name, strip_tags( $value ) );
}

add_filter( 'rest_query_vars', function ( $valid_vars ) {
    $valid_vars = array_merge( $valid_vars, array( 'meta_query' ) );
    return $valid_vars;
});

register_activation_hook( __FILE__, function () {

    // The space character after pagename= in the rewrite rules is necessary to
    // avoid triggering the broken 'verbose page match' check in
    // wp-includes/class-wp.php:parse_request. It's sufficient to defeat
    // the simplistic regexp there, and is trimmed by Wordpress during query
    // argument parsing.

    add_rewrite_rule(
        'docs/([^/]+)/([^/]+)/([^/]+)/([^/]+)',
        'index.php?post_type=documentation&pagename= $matches[1]-$matches[2]-$matches[3]/$matches[1]-$matches[2]-$matches[4]',
        'top'
    );

    add_rewrite_rule(
        'docs/([^/]+)/([^/]+)/([^/]+)',
        'index.php?post_type=documentation&pagename= $matches[1]-$matches[2]-$matches[3]',
        'top'
    );

    // Expensive, so flush on activation/deactivation only
    flush_rewrite_rules();
});

register_deactivation_hook( __FILE__, function () {
    // Expensive, so flush on activation/deactivation only
    flush_rewrite_rules();
});
