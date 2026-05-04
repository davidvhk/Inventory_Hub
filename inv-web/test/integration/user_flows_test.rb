require "test_helper"

class UserFlowsTest < ActionDispatch::IntegrationTest
  include Devise::Test::IntegrationHelpers

  setup do
    @user = User.create!(email: "test_integration@example.com", password: "password123", role: "user")
    @admin = User.create!(email: "admin_integration@example.com", password: "password123", role: "admin")
    @device = Device.create!(hostname: "test-device", device_id: "ID123", ip_address: "1.1.1.1")
  end

  test "unauthenticated user is redirected to login" do
    get root_path
    assert_redirected_to new_user_session_path
  end

  test "user can login and see dashboard but body has role-user class" do
    sign_in @user
    get root_path
    assert_response :success
    assert_select "body.role-user"
    # Button exists in HTML but hidden by CSS (which integration tests don't see)
    assert_select "button.btn-delete-device"
  end

  test "admin can login and see dashboard with role-admin class" do
    sign_in @admin
    get root_path
    assert_response :success
    assert_select "body.role-admin"
    assert_select "button.btn-delete-device"
  end

  test "admin can delete a device" do
    sign_in @admin
    assert_difference("Device.count", -1) do
      delete device_path(@device)
    end
    assert_redirected_to root_path
  end

  test "user cannot delete a device" do
    sign_in @user
    assert_no_difference("Device.count") do
      delete device_path(@device)
    end
    assert_redirected_to root_path
    assert_equal "Access denied. Admin only.", flash[:alert]
  end
end
