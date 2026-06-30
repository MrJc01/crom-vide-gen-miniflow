import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Layout from './Layout';
import Home from './Home';
import Editor from './Editor';
import VideosList from './VideosList';
import VideoDetail from './VideoDetail';
import Docs from './Docs';
import About from './About';
import MediaLibrary from './MediaLibrary';

export default function App() {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/editor/:id" element={<Editor />} />
          <Route path="/videos" element={<VideosList />} />
          <Route path="/videos/:id" element={<VideoDetail />} />
          <Route path="/docs" element={<Docs />} />
          <Route path="/about" element={<About />} />
          <Route path="/medias" element={<MediaLibrary />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  );
}
