terraform {
  # Only the root module can declare a Cloud configuration. OpenRecipe should emit a warning
  # about this child module Cloud declaration.
  cloud {
  }
}
