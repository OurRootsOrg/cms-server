<?php

/*====================================================
=            Create Options page and Menu            =
====================================================*/
if(!class_exists('OURROOTS')){
    class OURROOTS {
        public function __construct(){
			add_action( 'admin_menu', array($this, 'settings_page') );
			add_action( 'init', array($this, 'settings_boot'), 20 );
        	add_action( 'wp_enqueue_scripts', array($this, 'load_scripts') );
            add_shortcode( 'our-roots', array($this, 'shortcode_output') );
        }

		public function settings_page(){
			add_menu_page( 
				__('OurRootsDatabase', 'jwto'), 
				__('OurRootsDatabase', 'jwto'), 
				'manage_options', 
				OURROOTS_OPTIONS_SLUG, 
				array($this, 'settings_page_html'),
				// 'dashicons-editor-code'
				'dashicons-admin-generic'
			);
		}

		public function settings_page_html(){
			$jwto_settings = get_option('jwto_settings');
			?>
			<div class="wrap ourroots_options">
				<h2><?php echo esc_html('OurRootsDatabase', 'jwto'); ?></h2>
				<form method="post" action="">
					<table class="form-table">
						<tbody>
							<tr>
								<th scope="row"><?php echo __('Enable shortcode?', 'jwto'); ?></th>
								<td>
									<label>
										<input type="checkbox" name="jwto_enable" id="jwto_enable" <?php echo ((!empty($jwto_settings) && $jwto_settings['jwto_enable'] == "yes") ? 'checked' : '' ); ?>> Enable shortcode?</label>
								</td>
							</tr>
							<tr>
								<th scope="row"><?php echo __('Admin Domain', 'jwto'); ?></th>
								<td>
									<input type="text" name="jwto_admin_domain" id="jwto_admin_domain" value="<?php echo ((!empty($jwto_settings) && !empty($jwto_settings['jwto_admin_domain'])) ? $jwto_settings['jwto_admin_domain'] : '' ); ?>" class="regular-text">
									<p class="description"><?php echo __('Admin domain (e.g., https://db.ourroots.org)', 'jwto'); ?></p>
								</td>
							</tr>							
							<tr>
								<th scope="row"><?php echo __('Society ID', 'jwto'); ?></th>
								<td>
									<input type="number" name="jwto_society_id" id="jwto_society_id" value="<?php echo ((!empty($jwto_settings) && !empty($jwto_settings['jwto_society_id'])) ? $jwto_settings['jwto_society_id'] : '' ); ?>" min="1" max="9999" class="regular-text">
									<p class="description"><?php echo __('Society ID', 'jwto'); ?></p>
								</td>
							</tr>
							<tr>
								<th scope="row"><?php echo __('Secret key', 'jwto'); ?></th>
								<td>
									<input type="text" name="jwto_secret" id="jwto_secret" value="<?php echo ((!empty($jwto_settings) && !empty($jwto_settings['jwto_secret'])) ? $jwto_settings['jwto_secret'] : '' ); ?>" class="regular-text">
									<p class="description"><?php echo __('Secret key', 'jwto'); ?></p>
								</td>
							</tr>
							<tr>
								<th scope="row"><?php echo __('Token expiration days', 'jwto'); ?></th>
								<td>
									<input type="number" name="jwto_token_expire" id="jwto_token_expire" value="<?php echo ((!empty($jwto_settings) && !empty($jwto_settings['jwto_token_expire'])) ? $jwto_settings['jwto_token_expire'] : '7' ); ?>" min="1" max="14" class="regular-text">
									<p class="description"><?php echo __('Token expiration days', 'jwto'); ?></p>
								</td>
							</tr>
							<tr>
								<th scope="row"><?php echo __('Display names as surname, given', 'jwto'); ?></th>
								<td>
									<input type="text" name="jwto_surname_first" id="jwto_surname_first" value="<?php echo ((!empty($jwto_settings) && !empty($jwto_settings['jwto_surname_first'])) ? $jwto_settings['jwto_surname_first'] : '' ); ?>" class="regular-text">
									<p class="description"><?php echo __('True to display surname first; leave empty to display surname last', 'jwto'); ?></p>
								</td>
							</tr>
							<tr>
								<th scope="row"><?php echo __('Custom styles', 'jwto'); ?></th>
								<td>
									<textarea name="jwto_custom_css" id="jwto_custom_css" rows="5" cols="40"><?php echo ((!empty($jwto_settings) && !empty($jwto_settings['jwto_custom_css'])) ? $jwto_settings['jwto_custom_css'] : '' ); ?></textarea>
									<p class="description"><?php echo __('Custom styles', 'jwto'); ?></p>
								</td>
							</tr>
						</tbody>
					</table>
					<input type="hidden" name="_wpnonce" value="<?php echo wp_create_nonce('ourroots_nonce'); ?>">
					<input type="hidden" name="ourroots_action" value="save">
					<input type="submit" class="button button-primary" value="Save Changes">
				</form>
			</div>
			<?php
		}

		public function settings_boot(){
			if (
		        $_SERVER['REQUEST_METHOD'] === 'POST'
		        && isset( $_REQUEST["ourroots_action"] )
		        && wp_verify_nonce( $_REQUEST["_wpnonce"], "ourroots_nonce" )
		    ) {
				$jwto_settings = array();
				unset($_POST['_wpnonce']);
				unset($_POST['ourroots_action']);
				foreach($_POST as $key => $val){
					if($key != "jwto_custom_css"){
						$jwto_settings[sanitize_text_field($key)] = sanitize_text_field($val);
					} else {
						$jwto_settings[sanitize_textarea_field($key)] = sanitize_textarea_field($val);
					}
				}
				$jwto_settings['jwto_enable'] = (!isset($jwto_settings['jwto_enable'])) ? 'no' : 'yes';
		        update_option('jwto_settings', $jwto_settings);
		    }
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
			$jwto_settings = get_option('jwto_settings');
        	$jwto_enable = $jwto_settings['jwto_enable'];
        	if($jwto_enable != "yes"){
        		return;
        	}

        	$attributes = shortcode_atts( array(
		        'fields' => '',
		        'category' => '',
		        'collection' => '',
		    ), $atts );

        	$jwto_secret = $jwto_settings['jwto_secret'];
			$jwto_admin_domain = $jwto_settings['jwto_admin_domain'];
        	$jwto_society_id = $jwto_settings['jwto_society_id'];
        	$jwto_custom_css = $jwto_settings['jwto_custom_css'];
        	$jwto_token_expire = $jwto_settings['jwto_token_expire'];
			if (empty($jwto_token_expire)){
				$jwto_token_expire = '7';
			}
        	$jwto_surname_first = $jwto_settings['jwto_surname_first'];
			$date_modifier = ' +' . $jwto_token_expire . ' days';
			$exp_date = strtotime($date_modifier);
        	$expiration = date($exp_date);

			// $datetime = new DateTime();
			// $datetime->modify($date_modifier);
			// $expiration2 = $datetime->format('U');

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
				'jwto_token_expire' => $jwto_token_expire,
				'date_modifier' => $date_modifier,
				'exp_date' => $exp_date,
				'expiration' => $expiration,
				// 'expiration2' => $expiration2,
        		'jwt' => $token,
        		'fields' => $attributes['fields'],
        		'category' => $attributes['category'],
        		'collection' => $attributes['collection'],
        		'surnameFirst' => $jwto_surname_first,
				'adminDomain' => $jwto_admin_domain,
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
