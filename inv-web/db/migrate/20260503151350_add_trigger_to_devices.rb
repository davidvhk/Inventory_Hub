class AddTriggerToDevices < ActiveRecord::Migration[8.1]
  def up
    execute <<-SQL
      CREATE OR REPLACE FUNCTION notify_device_changes()
      RETURNS trigger AS $$
      BEGIN
        PERFORM pg_notify(
          'inventory_updates',
          json_build_object(
            'id', NEW.id,
            'action', LOWER(TG_OP)
          )::text
        );
        RETURN NEW;
      END;
      $$ LANGUAGE plpgsql;

      DROP TRIGGER IF EXISTS device_changes_trigger ON devices;
      CREATE TRIGGER device_changes_trigger
      AFTER INSERT OR UPDATE ON devices
      FOR EACH ROW EXECUTE FUNCTION notify_device_changes();

      CREATE OR REPLACE FUNCTION notify_device_deletion()
      RETURNS trigger AS $$
      BEGIN
        PERFORM pg_notify(
          'inventory_updates',
          json_build_object(
            'id', OLD.id,
            'action', 'delete'
          )::text
        );
        RETURN OLD;
      END;
      $$ LANGUAGE plpgsql;

      DROP TRIGGER IF EXISTS device_deletion_trigger ON devices;
      CREATE TRIGGER device_deletion_trigger
      AFTER DELETE ON devices
      FOR EACH ROW EXECUTE FUNCTION notify_device_deletion();
    SQL
  end

  def down
    execute <<-SQL
      DROP TRIGGER IF EXISTS device_changes_trigger ON devices;
      DROP TRIGGER IF EXISTS device_deletion_trigger ON devices;
      DROP FUNCTION IF EXISTS notify_device_changes();
      DROP FUNCTION IF EXISTS notify_device_deletion();
    SQL
  end
end
