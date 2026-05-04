Rails.application.routes.draw do
  devise_for :users
  resources :devices, only: [:index, :destroy]
  root "devices#index"
end
