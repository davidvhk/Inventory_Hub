Rails.application.routes.draw do
  get "devices/index"
  root "devices#index"
end
