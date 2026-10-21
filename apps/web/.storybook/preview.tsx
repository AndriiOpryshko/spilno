import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { Decorator, Preview } from '@storybook/react-vite';
import { useEffect } from 'react';

import '../src/index.css';

// The app toggles dark mode with a `.dark` class on <html>; do the same so the whole canvas switches.
function ThemeFrame({ dark, children }: { dark: boolean; children: React.ReactNode }) {
  useEffect(() => {
    document.documentElement.classList.toggle('dark', dark);
  }, [dark]);
  return (
    <QueryClientProvider client={new QueryClient()}>
      <div className="min-h-svh bg-background p-6 font-sans text-foreground">{children}</div>
    </QueryClientProvider>
  );
}

const withTheme: Decorator = (Story, { globals }) => (
  <ThemeFrame dark={globals.theme === 'dark'}>
    <Story />
  </ThemeFrame>
);

const preview: Preview = {
  globalTypes: {
    theme: {
      description: 'Color mode, mirrors the Light / Dark modes of the Figma Color collection',
      toolbar: { title: 'Theme', icon: 'mirror', items: ['light', 'dark'], dynamicTitle: true },
    },
  },
  initialGlobals: { theme: 'light' },
  decorators: [withTheme],
  parameters: {
    layout: 'fullscreen',
    // Accessibility violations fail the story tests, not just show up in the panel.
    a11y: { test: 'error' },
  },
};

export default preview;
