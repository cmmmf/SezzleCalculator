interface ErrorBannerProps {
  message: string | null
}

export function ErrorBanner({ message }: ErrorBannerProps) {
  if (!message) return null
  return (
    <p className="error-banner" role="alert">
      {message}
    </p>
  )
}
