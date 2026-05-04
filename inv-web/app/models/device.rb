class Device < ApplicationRecord
  include Turbo::Broadcastable
  validates :device_id, presence: true, uniqueness: true
end
