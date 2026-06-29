export default function Docs() {
  return (
    <div style={{ padding: '2rem', maxWidth: '800px', margin: '0 auto', overflowY: 'auto' }}>
      <h1 style={{ color: 'var(--accent)', marginBottom: '2rem' }}>Documentação Oficial</h1>
      
      <section style={{ marginBottom: '3rem' }}>
        <h2>1. Introdução aos Templates</h2>
        <p style={{ lineHeight: '1.6', color: 'var(--text-muted)' }}>
          Os templates são as fundações do seu vídeo. Ao acessar a aba de Início, você verá diversos templates criados previamente. Eles contém caixas delimitadoras e estruturas de texto e imagem.
        </p>
      </section>

      <section style={{ marginBottom: '3rem' }}>
        <h2>2. Modificando um Layout</h2>
        <p style={{ lineHeight: '1.6', color: 'var(--text-muted)' }}>
          No editor, você pode arrastar elementos e alterar seus tamanhos pelas bordas. Os elementos selecionados ficam destacados e suas propriedades aparecem no painel à direita (Inspector). Lá você pode mudar fontes, cores, z-index, entre outros.
        </p>
      </section>

      <section style={{ marginBottom: '3rem' }}>
        <h2>3. A Nova Tela de Pré-Renderizador</h2>
        <p style={{ lineHeight: '1.6', color: 'var(--text-muted)' }}>
          Ao invés de definir o conteúdo final (o texto ou arquivo exato) na etapa de design, você clica em <strong>🚀 Render</strong>. Isso abrirá um modal solicitando que você preencha as "variáveis" do template com os arquivos mp4, imagens e textos reais daquele projeto! A duração será ajustada automaticamente baseada nos vídeos inseridos.
        </p>
      </section>
    </div>
  );
}
