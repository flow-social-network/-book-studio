import Link from 'next/link';
import {ArrowRight,BookOpen} from 'lucide-react';

export default function EditorIndexRoute() {
  return <div className="mx-auto max-w-3xl"><h1 className="text-3xl font-semibold tracking-tight">Editor de manuscrito</h1><p className="mt-2 text-sm leading-6 text-slate-500">Cada livro possui um editor próprio, ligado ao projeto persistido.</p><section className="mt-6 rounded-2xl border border-slate-200 bg-white p-6"><BookOpen size={26} className="text-blue-700"/><h2 className="mt-4 font-semibold">Selecione um projeto</h2><p className="mt-2 text-sm leading-6 text-slate-500">Abra um projeto existente ou crie um novo para editar seu conteúdo.</p><div className="mt-5 flex flex-wrap gap-3"><Link href="/app/projetos" className="inline-flex items-center gap-2 rounded-xl bg-blue-600 px-4 py-2.5 text-sm font-semibold text-white">Meus projetos <ArrowRight size={16}/></Link><Link href="/app/projetos/novo" className="rounded-xl border border-slate-200 px-4 py-2.5 text-sm font-semibold">Criar livro</Link></div></section></div>;
}
