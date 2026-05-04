require "test_helper"

class DevicesControllerTest < ActionDispatch::IntegrationTest
  include Devise::Test::IntegrationHelpers

  setup do
    @user = User.create!(email: "test_controller@example.com", password: "password123", role: "user")
  end

  test "should get index when signed in" do
    sign_in @user
    get devices_url
    assert_response :success
  end

  test "should redirect root to sign in when not signed in" do
    get root_url
    assert_redirected_to new_user_session_path
  end
end
