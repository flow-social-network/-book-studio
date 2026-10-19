import {Suspense} from 'react';
import ConfirmPasswordResetPage from '@/components/features/account/ConfirmPasswordResetPage';
export default function ConfirmPasswordResetRoute(){return <Suspense fallback={<main className="p-8 text-sm text-slate-500">Carregando formulário…</main>}><ConfirmPasswordResetPage/></Suspense>}
