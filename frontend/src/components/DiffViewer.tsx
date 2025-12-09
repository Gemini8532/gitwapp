import React, { useMemo } from 'react';
import { parseDiff, Diff, Hunk } from 'react-diff-view';
import 'react-diff-view/style/index.css';

interface DiffViewerProps {
    diffText: string;
    viewType?: 'unified' | 'split';
}

export const DiffViewer: React.FC<DiffViewerProps> = ({ diffText, viewType: propViewType }) => {
    const files = useMemo(() => parseDiff(diffText), [diffText]);

    // Custom hook to detect mobile view if viewType isn't forced
    const isMobile = useMediaQuery('(max-width: 768px)');
    const viewType = propViewType || (isMobile ? 'unified' : 'split');

    if (!diffText) return <div className="text-gray-500 italic">No content to diff</div>;

    return (
        <div className="diff-viewer-container">
            {files.map((file) => (
                <div key={file.oldPath + file.newPath} className="mb-4 border rounded overflow-hidden">
                    <div className="bg-gray-100 dark:bg-gray-800 p-2 text-sm font-mono border-b dark:border-gray-700">
                        {file.newPath || file.oldPath}
                    </div>
                    <div className="overflow-x-auto text-xs">
                        <Diff viewType={viewType} diffType={file.type} hunks={file.hunks}>
                            {(hunks: any[]) => hunks.map((hunk) => <Hunk key={hunk.content} hunk={hunk} />)}
                        </Diff>
                    </div>
                </div>
            ))}
        </div>
    );
};

// Simple media query hook
const useMediaQuery = (query: string) => {
    const [matches, setMatches] = React.useState(window.matchMedia(query).matches);

    React.useEffect(() => {
        const media = window.matchMedia(query);
        if (media.matches !== matches) {
            setMatches(media.matches);
        }
        const listener = () => setMatches(media.matches);
        media.addEventListener('change', listener);
        return () => media.removeEventListener('change', listener);
    }, [query, matches]);

    return matches;
}
