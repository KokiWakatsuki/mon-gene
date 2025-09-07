import React from 'react';

export default function Header() {
  return (
    <header className="flex items-center justify-between mb-5">
      <div className="flex items-center gap-2.5">
        <div 
          className="w-8 h-8 bg-mongene-blue rounded-lg" 
          aria-hidden="true"
        />
        <div className="font-extrabold text-mongene-blue text-lg">
          Mongene
        </div>
      </div>
      <div 
        className="w-10 h-10 bg-mongene-border rounded-full" 
        aria-hidden="true"
      />
    </header>
  );
}
