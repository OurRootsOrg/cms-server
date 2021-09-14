<?php

/*====================================================
=            Create Options page and Menu            =
====================================================*/
if(!class_exists('OURROOTS')){
    class OURROOTS {
        public function __construct(){
        	add_action( 'wp_enqueue_scripts', array($this, 'load_scripts') );
            add_shortcode( 'our-roots', array($this, 'shortcode_output') );
        }

        public function load_scripts(){
            global $post;
            if ( isset($post->post_content) && strpos($post->post_content, '[our-roots') !== false ) {
                wp_enqueue_style('jwto-fonts', OURROOTS_URL . '/css/ourroots.css', array(), time());
                wp_enqueue_style('jwto-icons', OURROOTS_URL . '/css/materialdesignicons.min.css', array(), time());
                wp_enqueue_style('jwto-css-chunk-vendors', OURROOTS_URL . '/css/chunk-vendors.css', array(), time());
                wp_enqueue_style('jwto-css-app', OURROOTS_URL . '/css/app.css', array(), time());
                wp_enqueue_script('jwto-js-chunk-vendors', OURROOTS_URL . '/js/chunk-vendors.js', array(), time(), true);
                wp_enqueue_script('jwto-js-app', OURROOTS_URL . '/js/app.js', array(), time(), true);
            }
        }

        public function base64url_encode($data) { 
		    return rtrim(strtr(base64_encode($data), '+/', '-_'), '='); 
		} 

		public function base64url_decode($data) { 
		    return base64_decode(str_pad(strtr($data, '-_', '+/'), strlen($data) % 4, '=', STR_PAD_RIGHT)); 
		}

        public function shortcode_output($atts){
        	ob_start();
        	$jwto_enable = get_option('jwto_enable');
        	if($jwto_enable != "on"){
        		return;
        	}

        	$attributes = shortcode_atts( array(
		        'fields' => '',
		        'category' => '',
		        'collection' => '',
		    ), $atts );

        	$jwto_secret = get_option('jwto_secret');
        	$jwto_society_id = get_option('jwto_society_id');
        	$jwto_custom_css = get_option('jwto_custom_css');
        	$jwto_token_expire = get_option('jwto_token_expire');
        	$expiration = date(strtotime(' +' . $jwto_token_expire . ' days'));

        	$user_id = 0;
        	if(is_user_logged_in()){
        		$user_id = get_current_user_id();
        	}
        	$expiration = (int)$expiration;
        	$society_id = (int)$jwto_society_id;
        	$subject = "{$society_id}_{$user_id}";

        	//build the headers
            $headers = ['alg'=>'HS256','typ'=>'JWT'];
            $headers_encoded = $this->base64url_encode(json_encode($headers));

            //build the payload
            $payload = ['sub' => $subject, 'exp' => $expiration];
            $payload_encoded = $this->base64url_encode(json_encode($payload));

            //build the signature
            $key = $jwto_secret;
            $signature = hash_hmac('SHA256',"$headers_encoded.$payload_encoded",$key, true);
            $signature_encoded = $this->base64url_encode($signature);

            //build and return the token
            $token = "$headers_encoded.$payload_encoded.$signature_encoded";

        	$js_array = array(
        		'jwt' => $token,
        		'fields' => $attributes['fields'],
        		'category' => $attributes['category'],
        		'collection' => $attributes['collection'],
        		'societyId' => $society_id,
        		'images_directory' => OURROOTS_URL . '/',
        	);
        	?>
			<!-- Custom styles can go here. In this example we are overriding a style from our wordpress theme. -->
			<style type="text/css"><?php echo esc_html($jwto_custom_css); ?></style>
			<div id="app"></div>
			<script type="text/javascript">
				window.ourroots = JSON.parse('<?php echo json_encode($js_array) ?>');
			</script>
        	<?php
        	return ob_get_clean();
        }
    }
    new OURROOTS();
}
?>
