require "test_helper"

class DeviceTest < ActiveSupport::TestCase
  test "should be valid with attributes" do
    device = Device.new(
      hostname: "test-host",
      ip_address: "127.0.0.1",
      device_id: "test-id-123"
    )
    assert device.valid?
  end

  test "should save device" do
    device = Device.new(
      hostname: "test-host",
      ip_address: "127.0.0.1",
      device_id: "test-id-123"
    )
    assert device.save
  end
end
