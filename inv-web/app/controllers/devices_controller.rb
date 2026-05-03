class DevicesController < ApplicationController
  def index
    @devices = Device.all.order(updated_at: :desc)
  end
end
