import type { Metadata, Viewport } from 'next';
import './globals.css';

const siteName = 'Book Studio';
const description = 'Crie, escreva, ilustre, revise, diagrama e prepare livros para publicação num único estúdio editorial com IA.';

export const metadata: Metadata = {
  title: { default: 'Book Studio — Seu estúdio editorial com IA', template: '%s | Book Studio' },
  description,
  applicationName: siteName,
  generator: 'Next.js',
  keywords: ['Book Studio', 'criação de livros', 'escrita com IA', 'editorial', 'e-book', 'EPUB', 'PDF', 'publicação de livros', 'diagramação'],
  authors: [{ name: 'Book Studio' }],
  creator: 'Book Studio',
  publisher: 'Book Studio',
  category: 'books',
  alternates: { languages: { 'pt-BR': '/' } },
  icons: {
    icon: [{ url: '/icon.svg', type: 'image/svg+xml' }],
    shortcut: '/icon.svg',
    apple: '/icon.svg',
  },
  manifest: '/site.webmanifest',
  openGraph: {
    type: 'website',
    locale: 'pt_BR',
    siteName,
    title: 'Book Studio — Da ideia ao livro publicado',
    description,
    images: [{ url: '/opengraph-image.svg', width: 1200, height: 360, alt: 'Book Studio — Crie, escreva, ilustre, diagrama e publique' }],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Book Studio — Seu estúdio editorial com IA',
    description,
    images: [{ url: '/opengraph-image.svg', alt: 'Book Studio' }],
  },
  robots: { index: true, follow: true, googleBot: { index: true, follow: true, 'max-image-preview': 'large', 'max-snippet': -1, 'max-video-preview': -1 } },
  appleWebApp: { capable: true, title: siteName, statusBarStyle: 'default' },
};

export const viewport: Viewport = { width: 'device-width', initialScale: 1, themeColor: '#0B5FFF', colorScheme: 'light' };

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return <html lang="pt-BR"><body>{children}</body></html>;
}
