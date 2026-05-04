# Create a default admin user
User.find_or_create_by!(email: 'admin@inventoryhub.com') do |u|
  u.password = 'password123'
  u.role = 'admin'
end

# Create a regular user
User.find_or_create_by!(email: 'user@inventoryhub.com') do |u|
  u.password = 'password123'
  u.role = 'user'
end

puts "Seed data created: admin@inventoryhub.com and user@inventoryhub.com (password: password123)"
