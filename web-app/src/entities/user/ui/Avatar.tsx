import { useEffect, useState } from "react";

interface AvatarProps {
  src?: string;
  initials: string;
  className: string;
  alt?: string;
}

export function Avatar({ src, initials, className, alt = "" }: AvatarProps) {
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    setFailed(false);
  }, [src]);

  return (
    <span className={className}>
      {src && !failed ? <img src={src} alt={alt} onError={() => setFailed(true)} /> : initials}
    </span>
  );
}
