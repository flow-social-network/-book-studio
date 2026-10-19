import {Suspense} from 'react';
import GoogleDrivePage from '@/components/features/integrations/GoogleDrivePage';
export default function GoogleDriveRoute(){return <Suspense fallback={<main className="p-8 text-sm text-slate-500">Carregando Google Drive…</main>}><GoogleDrivePage/></Suspense>}
