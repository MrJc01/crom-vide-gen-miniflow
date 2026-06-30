# Interface Web — Guia do Usuário

A interface web é uma aplicação React + Vite que permite criar, editar e renderizar vídeos visualmente.

## Páginas

### 🏠 Início (`/`)

Grid com todos os templates JSON disponíveis na pasta `templates/examples/`. Clique em qualquer template para abrir o editor visual.

- **Pesquisa:** Campo de busca para filtrar templates pelo nome.

### ✏️ Editor (`/editor/:id`)

Editor visual interativo com três painéis:

1. **Canvas (Centro):** Preview em tempo real do frame atual do card selecionado.
2. **Inspector (Direita):** Painel de propriedades para editar os elementos do card (cores, textos, posições, tamanhos).
3. **Timeline (Inferior):** Lista horizontal de todos os cards (cenas) do template.

#### Ações do Editor

| Botão | Ação |
|-------|------|
| 🎬 Render | Abre o modal de pré-renderização para configurar e disparar o processamento |
| ❐ Clonar | Cria uma cópia idêntica do card selecionado na timeline |
| ← → Mover | Reordena a posição dos cards na sequência do vídeo |
| × Deletar | Remove o card selecionado |

### 🎥 Meus Vídeos (`/videos`)

Dashboard de todos os vídeos já renderizados, com status em tempo real:

- **✅ CONCLUÍDO** — Vídeo pronto para download.
- **⏳ RENDERIZANDO** — Processamento em andamento (atualiza automaticamente a cada 2s).
- **❌ ERRO** — Falha na renderização (mostra detalhes do erro).

Ações por vídeo: Arquivar, Desarquivar, Excluir.

### 🎥 Detalhes do Vídeo (`/videos/:id`)

Player HTML5 do vídeo renderizado, com:

- **Tempo de renderização** (ex: `Renderizado em 12.50s`)
- **Blueprint JSON** — Seção recolhível contendo o JSON exato que gerou o vídeo, com botão de cópia rápida.
- **Download MP4** — Link direto para baixar o arquivo.
- **Editar e Gerar Novo** — Abre o editor com o mesmo template para iterações rápidas.

### 📂 Mídias (`/medias`)

Gerenciador de assets organizado por subpastas de `tmp/uploads/`.

- **Barra lateral:** Navegação por pastas com contagem de arquivos.
- **Filtros:** Por tipo (Vídeos, Imagens, Áudio) e por nome.
- **Preview:** Thumbnails de imagens e players para vídeos.
- **Ações:** Copiar caminho local, abrir URL pública, enviar nova mídia.

### 📚 Documentação (`/docs`)

Página informativa com instruções de uso.

### ℹ️ Sobre (`/about`)

Informações sobre o projeto.

---

## Modal de Pré-Renderização

Ao clicar em **Renderizar** no editor, o modal permite:

1. **Selecionar Mídias:** Cada slot de vídeo/imagem tem um botão `🔍 Selecionar Mídia` que abre a Galeria de Mídias integrada.
2. **Editar Textos:** Campos de texto editáveis para personalizar os conteúdos antes do render final.
3. **Configurar Duração:** Modo automático (detecta do vídeo via `ffprobe`) ou manual (em segundos).
4. **Trilha Sonora Global:** Selecionar um áudio da biblioteca de mídias.
5. **Aceleração GPU:** Checkbox para habilitar NVENC (com fallback automático).
6. **Qualidade JPEG:** Seletor (Alta, Média, Baixa) que controla o I/O de disco dos frames temporários.

---

## Acesso por Rede Local

A interface detecta automaticamente o IP do servidor via `window.location.hostname`. Se o Vite estiver rodando com `--host`, qualquer dispositivo na rede local pode acessar:

```
http://192.168.X.X:5173/
```

Os previews de mídias e as chamadas de API funcionarão transparentemente sem configuração adicional.
