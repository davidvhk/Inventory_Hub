require "test_helper"

class DevicesControllerTest < ActionDispatch::IntegrationTest
  test "should get index" do
    get devices_index_url
    assert_response :success
  end

  test "should get root" do
    get root_url
    assert_response :success
  end
end
