class DevicesController < ApplicationController
  before_action :authenticate_user!
  before_action :set_device, only: [:destroy]
  before_action :authorize_admin!, only: [:destroy]

  def index
    @devices = Device.all.order(updated_at: :desc)
  end

  def destroy
    @device.destroy
    respond_to do |format|
      format.html { redirect_to root_path, notice: "Device deleted." }
      format.turbo_stream
    end
  end

  private

  def set_device
    @device = Device.find(params[:id])
  end

  def authorize_admin!
    unless current_user.role == 'admin'
      redirect_to root_path, alert: "Access denied. Admin only."
    end
  end
end
