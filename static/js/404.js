function createParticles() {
    const particles = document.querySelector('.particles');
    for (let i = 0; i < 20; i++) {
        const particle = document.createElement('div');
        particle.className = 'particle';
        particle.style.width = Math.random() * 6 + 2 + 'px';
        particle.style.height = particle.style.width;
        particle.style.left = Math.random() * 100 + '%';
        particle.style.animationDelay = Math.random() * 6 + 's';
        particle.style.animationDuration = (Math.random() * 4 + 4) + 's';
        
        if (i % 2 === 0) {
            particle.style.background = 'rgba(115, 217, 217, 0.2)';
        } else {
            particle.style.background = 'rgba(242, 82, 96, 0.2)';
        }
        
        particles.appendChild(particle);
    }
}

document.addEventListener('mousemove', (e) => {
    const cursor = document.querySelector('.cursor');
    if (!cursor) {
        const newCursor = document.createElement('div');
        newCursor.className = 'cursor';
        newCursor.style.cssText = `
            position: fixed;
            width: 20px;
            height: 20px;
            background: rgba(115, 217, 217, 0.3);
            border: 2px solid rgba(242, 82, 96, 0.5);
            border-radius: 50%;
            pointer-events: none;
            z-index: 9999;
            transition: all 0.1s ease;
        `;
        document.body.appendChild(newCursor);
    }
    
    const cursorElement = document.querySelector('.cursor');
    cursorElement.style.left = e.clientX - 10 + 'px';
    cursorElement.style.top = e.clientY - 10 + 'px';
});

function randomNotFoundPhrases() {
    const notFoundPhrases = window.pageData;

    document.getElementById("notFoundPhrase").innerHTML = notFoundPhrases[Math.floor(Math.random() * 4)];
}
        
randomNotFoundPhrases();
createParticles();