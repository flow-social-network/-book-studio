import type { Metadata, Viewport } from 'next';
import './globals.css';
export const metadata: Metadata = { title: { default: 'Book Studio — Seu estúdio editorial com IA', template: '%s | Book Studio' }, description: 'Planeje, escreva, ilustre, edite e prepare livros para publicação em um único estúdio editorial.', applicationName: 'Book Studio' };
export const viewport: Viewport = { width: 'device-width', initialScale: 1, themeColor: '#ffffff' };
export default function RootLayout({children}:{children:React.ReactNode}) { return <html lang="pt-BR"><body>{children}</body></html>; }
