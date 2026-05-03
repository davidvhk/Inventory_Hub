class CreateDevices < ActiveRecord::Migration[8.1]
  def change
    create_table :devices do |t|
      t.datetime :inventory_timestamp
      t.string :hostname
      t.string :ip_address
      t.string :os
      t.string :uuid
      t.string :device_id, index: { unique: true }

      t.timestamps
    end
  end
end
