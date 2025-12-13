document.addEventListener('DOMContentLoaded', () => {
  App.init();
});

/* =========================================
   Configuration & Constants
   ========================================= */
const CONFIG = {
  bgShapes: [
    "data:image/svg+xml,%3Csvg viewBox='0 0 243 122' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M237.266,121.4l-231.731,0c-3.057,-0 -5.534,-2.478 -5.534,-5.534l0,-110.331c0,-3.057 2.478,-5.534 5.534,-5.534l231.731,0c3.057,0 5.534,2.478 5.534,5.534l0,110.331c-0,3.057 -2.478,5.534 -5.534,5.534Z' fill='black'/%3E%3C/svg%3E",
    "data:image/svg+xml,%3Csvg viewBox='0 0 243 122' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M64.833,116.376c-0.25,2.838 -2.626,5.014 -5.475,5.014c-14.469,0.01 -53.823,0.01 -53.823,0.01c-3.057,-0 -5.534,-2.478 -5.534,-5.534l0,-110.331c0,-3.057 2.478,-5.534 5.534,-5.534l231.731,0c3.057,0 5.534,2.478 5.534,5.534l0,110.331c-0,3.057 -2.478,5.534 -5.534,5.534c-0,0 -39.356,0 -53.823,0c-2.854,-0 -5.235,-2.18 -5.485,-5.023c-2.537,-28.993 -26.909,-51.764 -56.557,-51.764c-29.649,0 -54.021,22.771 -56.567,51.763Z' fill='black'/%3E%3C/svg%3E",
    "data:image/svg+xml,%3Csvg viewBox='0 0 113 57' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M5.506,56.787c-1.565,-0 -3.056,-0.666 -4.1,-1.831c-1.044,-1.165 -1.544,-2.72 -1.373,-4.275c3.059,-28.474 27.192,-50.681 56.471,-50.681c29.279,0 53.412,22.207 56.462,50.682c0.17,1.553 -0.328,3.105 -1.371,4.268c-1.043,1.163 -2.531,1.828 -4.093,1.828c-19.479,0.009 -82.516,0.009 -101.995,0.009Z' fill='black'/%3E%3C/svg%3E",
    "data:image/svg+xml,%3Csvg viewBox='0 0 122 122' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M115.866,0c3.057,0 5.534,2.478 5.534,5.534l0,110.331c-0,3.057 -2.478,5.534 -5.534,5.534l-110.331,0c-3.057,-0 -5.534,-2.478 -5.534,-5.534l-0,-110.331c0,-3.057 2.478,-5.534 5.534,-5.534l110.331,0Z' fill='black'/%3E%3C/svg%3E",
    "data:image/svg+xml,%3Csvg viewBox='0 0 114 114' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M111.952,104.126c1.583,1.583 2.056,3.963 1.2,6.031c-0.857,2.068 -2.875,3.416 -5.113,3.416l-102.505,-0c-3.057,-0 -5.534,-2.478 -5.534,-5.534l0,-102.505c-0,-2.238 1.348,-4.256 3.416,-5.113c2.068,-0.857 4.448,-0.383 6.031,1.2c22.652,22.652 79.852,79.852 102.505,102.505Z' fill='black'/%3E%3C/svg%3E",
    "data:image/svg+xml,%3Csvg viewBox='0 0 225 118' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M7.968,117.416c-1.707,0.853 -3.734,0.762 -5.357,-0.241c-1.623,-1.003 -2.611,-2.775 -2.611,-4.684l-0,-106.985c0,-3.041 2.465,-5.506 5.506,-5.506c28.489,-0 159.911,-0 213.97,-0c2.554,-0 4.772,1.756 5.359,4.241c0.587,2.485 -0.612,5.048 -2.896,6.19c-51.054,25.527 -180.549,90.274 -213.97,106.985Z' fill='black'/%3E%3C/svg%3E",
    "data:image/svg+xml,%3Csvg viewBox='0 0 114 114' xmlns='http://www.w3.org/2000/svg'%3E%3Ccircle cx='56.787' cy='56.787' r='56.787' fill='black'/%3E%3C/svg%3E",
  ],
  colors: ['color-green', 'color-yellow', 'color-blue', 'color-pink'],
  mockData: {
    units: ['二次関数', '図形の相似', '円の性質', '三平方の定理', '数と式', '二次方程式'],
    sources: ['2025年 第1回', '2024年 第3回', '2024年 追試', '共通テスト 2023', 'オリジナル'],
    statuses: ['未チェック', 'チェック済み']
  }
};

/* =========================================
   Application Logic
   ========================================= */
const App = {
  init() {
    this.initBackground();
    this.initProblemList();
    this.initUIInteractions();
    this.setupFileUploaders();
    this.initTagSystem();
    this.initGenerationProcess();
    this.initFilterLogic();
  },

  /* 1. 背景生成ロジック */
  initBackground() {
    const container = document.getElementById('bg-container');
    if (!container) return;
    
    // 断片化を避けるためドキュメントフラグメントを使用
    const fragment = document.createDocumentFragment();
    const cols = 6;
    const rows = 5;
    const xStep = 100 / cols;
    const yStep = 100 / rows;

    for (let r = 0; r < rows; r++) {
      for (let c = 0; c < cols; c++) {
        if (Math.random() > 0.5) continue;
        
        const div = document.createElement('div');
        div.classList.add('bg-shape');
        
        const shapeUrl = CONFIG.bgShapes[Math.floor(Math.random() * CONFIG.bgShapes.length)];
        div.style.maskImage = `url("${shapeUrl}")`;
        div.style.webkitMaskImage = `url("${shapeUrl}")`;
        
        const colorClass = CONFIG.colors[Math.floor(Math.random() * CONFIG.colors.length)];
        div.classList.add(colorClass);
        
        const baseX = c * xStep;
        const baseY = r * yStep;
        div.style.left = `${baseX + Math.random() * (xStep * 0.6)}%`;
        div.style.top = `${baseY + Math.random() * (yStep * 0.6)}%`;
        
        const sizeBase = 4 + Math.random() * 4;
        div.style.width = `${sizeBase}vw`;
        div.style.height = `${sizeBase}vw`;
        div.style.transform = `rotate(${Math.random() * 360}deg)`;
        
        fragment.appendChild(div);
      }
    }
    container.appendChild(fragment);
  },

  /* 2. ダミー問題リスト生成 */
  initProblemList() {
    const problemList = document.querySelector('.problem-list');
    if (!problemList) return;

    problemList.innerHTML = '';
    const pick = (arr) => arr[Math.floor(Math.random() * arr.length)];
    const fragment = document.createDocumentFragment();

    for (let i = 1; i <= 12; i++) {
      const unit = pick(CONFIG.mockData.units);
      const source = pick(CONFIG.mockData.sources);
      const status = pick(CONFIG.mockData.statuses);
      const statusClass = status === '未チェック' ? 'tag-status is-unchecked' : 'tag-status';

      const card = document.createElement('div');
      card.className = 'problem-card';
      card.innerHTML = `
        <div class="problem-card-content">
          <h3>問題 ${i}</h3>
          <div class="problem-tags">
            <span class="tag-badge tag-unit">${unit}</span>
            <span class="tag-badge tag-source">${source}</span>
            <span class="tag-badge ${statusClass}">${status}</span>
          </div>
          <div class="problem-card-body">ここに「${unit}」に関する問題文が入ります。出典は${source}です。</div>
        </div>
        <div class="problem-card-actions">
          <button class="btn btn-blue" style="padding: 8px 16px;">プレビュー</button>
          <button class="btn" style="padding: 8px 16px; background: var(--bg-hover); color: var(--text-main);">印刷</button>
        </div>
      `;
      fragment.appendChild(card);
    }
    problemList.appendChild(fragment);
  },

  /* 3. UI動作 (Modal, Tabs, Accordion) */
  initUIInteractions() {
    // アカウントモーダル
    const accountToggle = document.getElementById('accountToggle');
    const accountModal = document.getElementById('accountModal');
    
    if (accountToggle && accountModal) {
      accountToggle.addEventListener('click', (e) => {
        e.stopPropagation();
        const isHidden = !accountModal.hidden;
        accountModal.hidden = isHidden;
        accountToggle.setAttribute('aria-expanded', !isHidden);
      });
      document.addEventListener('click', (e) => {
        if (!accountModal.hidden && !accountModal.contains(e.target) && e.target !== accountToggle) {
          accountModal.hidden = true;
          accountToggle.setAttribute('aria-expanded', 'false');
        }
      });
    }

    // カテゴリーフィルターのアコーディオン
    const accordionBtn = document.getElementById('categoryAccordionBtn');
    const accordionBody = document.getElementById('categoryAccordionBody');
    const accordionContainer = document.querySelector('.filter-accordion');

    if (accordionBtn && accordionBody && accordionContainer) {
      accordionBtn.addEventListener('click', () => {
        const willOpen = !accordionContainer.classList.contains('is-open');
        accordionContainer.classList.toggle('is-open', willOpen);
        accordionBtn.setAttribute('aria-expanded', willOpen);
        
        if (willOpen) {
          accordionBody.style.maxHeight = accordionBody.scrollHeight + "px";
        } else {
          accordionBody.style.maxHeight = accordionBody.scrollHeight + "px"; // trigger reflow
          setTimeout(() => { accordionBody.style.maxHeight = "0px"; }, 10);
        }
      });
    }

    // 単元の「もっと見る」
    const toggleUnitsBtn = document.getElementById('toggleUnitsBtn');
    const moreUnitsArea = document.getElementById('moreUnitsArea');
    if (toggleUnitsBtn && moreUnitsArea) {
      toggleUnitsBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        const isHidden = moreUnitsArea.style.display === 'none';
        
        moreUnitsArea.style.display = isHidden ? 'block' : 'none';
        toggleUnitsBtn.querySelector('span').textContent = isHidden ? '閉じる' : 'もっと見る';
        toggleUnitsBtn.classList.toggle('is-open', isHidden);

        // 親アコーディオンの高さ再計算
        if (accordionBody && accordionContainer.classList.contains('is-open')) {
          requestAnimationFrame(() => {
            accordionBody.style.maxHeight = accordionBody.scrollHeight + "px";
          });
        }
      });
    }

    // メインタブ
    const mainTabButtons = document.querySelectorAll('.main-tab');
    const mainTabPanels = document.querySelectorAll('.tab-panel');
    mainTabButtons.forEach(button => {
      button.addEventListener('click', () => {
        const targetTabId = button.getAttribute('data-tab');
        mainTabButtons.forEach(btn => btn.classList.remove('active'));
        mainTabPanels.forEach(panel => panel.classList.remove('active'));
        
        button.classList.add('active');
        const targetPanel = document.getElementById(targetTabId);
        if(targetPanel) targetPanel.classList.add('active');
      });
    });
  },

  /* 4. 検索フィルタリング */
  initFilterLogic() {
    const keywordInput = document.getElementById('keywordInput');
    const searchBtn = document.getElementById('keywordSearchBtn');
    if (!keywordInput || !searchBtn) return;

    const executeSearch = () => {
      const keyword = keywordInput.value.toLowerCase().trim();
      const cards = document.querySelectorAll('.problem-card');

      cards.forEach(card => {
        const text = card.textContent.toLowerCase();
        const isVisible = !keyword || text.includes(keyword);
        card.style.display = isVisible ? '' : 'none';
        if(isVisible) {
          card.style.opacity = '0';
          setTimeout(() => card.style.opacity = '1', 50);
        }
      });
    };

    searchBtn.addEventListener('click', executeSearch);
    keywordInput.addEventListener('keypress', (e) => {
      if (e.key === 'Enter') executeSearch();
    });
  },

  /* 5. ファイルアップロード */
  setupFileUploaders() {
    const setup = (areaId, inputId, listId) => {
      const uploadArea = document.getElementById(areaId);
      const fileInput = document.getElementById(inputId);
      const previewList = document.getElementById(listId);
      let uploadedFiles = [];

      if (!uploadArea || !fileInput) return;

      // Click to upload
      uploadArea.addEventListener('click', () => fileInput.click());

      // Drag & Drop events
      ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(evt => {
        uploadArea.addEventListener(evt, (e) => { e.preventDefault(); e.stopPropagation(); });
      });
      
      ['dragenter', 'dragover'].forEach(evt => uploadArea.addEventListener(evt, () => uploadArea.classList.add('dragover')));
      ['dragleave', 'drop'].forEach(evt => uploadArea.addEventListener(evt, () => uploadArea.classList.remove('dragover')));

      uploadArea.addEventListener('drop', (e) => addFiles(Array.from(e.dataTransfer.files)));
      fileInput.addEventListener('change', () => {
        addFiles(Array.from(fileInput.files));
        fileInput.value = '';
      });

      const addFiles = (files) => {
        if (!files.length) return;
        uploadedFiles = [...uploadedFiles, ...files];
        renderPreviews();
      };

      const renderPreviews = () => {
        previewList.innerHTML = '';
        previewList.hidden = (uploadedFiles.length === 0);
        
        uploadedFiles.forEach((file, index) => {
          const card = document.createElement('div');
          card.className = 'file-card';
          let contentHtml = '<span class="file-type-icon">FILE</span>';
          
          if (file.type.startsWith('image/') && !file.name.toLowerCase().endsWith('.heic')) {
            contentHtml = `<img src="${URL.createObjectURL(file)}" alt="preview">`;
          } else if (file.type === 'application/pdf') {
            contentHtml = `<span class="file-type-icon pdf">PDF</span>`;
          }
          
          card.innerHTML = `
            <div class="file-card-thumb">${contentHtml}</div>
            <button type="button" class="btn-remove-file" aria-label="削除">×</button>
            <div class="file-card-name" title="${file.name}">${file.name}</div>
          `;
          
          // イベントデリゲーションではなくクロージャで個別に設定(要素数が少ないため許容)
          card.querySelector('.btn-remove-file').addEventListener('click', (e) => {
            e.stopPropagation();
            uploadedFiles.splice(index, 1);
            renderPreviews();
          });
          previewList.appendChild(card);
        });
      };
    };

    setup('uploadAreaProblem', 'fileInputProblem', 'previewListProblem');
    setup('uploadAreaAnswer', 'fileInputAnswer', 'previewListAnswer');
  },

  /* 6. タグシステム */
  initTagSystem() {
    const tagInput = document.getElementById('tagSearchInput');
    const suggestionsBox = document.getElementById('tagSuggestions');
    const activeTagsArea = document.getElementById('activeTagsArea');
    const checkboxes = document.querySelectorAll('.checkbox-list input[type="checkbox"]');
    const btnAddSource = document.getElementById('btnAddSource');
    const selYear = document.getElementById('selYear');
    const selExam = document.getElementById('selExam');

    const state = {
      checked: new Map(), // Map<label, inputElement>
      custom: new Set()   // Set<string>
    };

    const searchDatabase = Array.from(checkboxes).map(input => ({
      label: input.dataset.label,
      dom: input
    }));

    const renderTags = () => {
      activeTagsArea.innerHTML = '';
      
      // チェックボックス由来のタグ
      state.checked.forEach((inputEl, label) => {
        const category = inputEl.dataset.category || 'default';
        activeTagsArea.appendChild(createTagElement(label, category, () => {
          inputEl.checked = false;
          inputEl.dispatchEvent(new Event('change'));
        }));
      });

      // 手動追加のタグ
      state.custom.forEach(label => {
        activeTagsArea.appendChild(createTagElement(label, 'source', () => {
          state.custom.delete(label);
          renderTags();
        }));
      });
    };

    const createTagElement = (label, category, removeFn) => {
      const tagEl = document.createElement('div');
      tagEl.className = `tag-chip tag-${category}`;
      tagEl.innerHTML = `<span>${label}</span><button type="button" class="tag-close-btn">×</button>`;
      tagEl.querySelector('.tag-close-btn').addEventListener('click', removeFn);
      return tagEl;
    };

    checkboxes.forEach(input => {
      input.addEventListener('change', () => {
        const label = input.dataset.label;
        if (input.checked) state.checked.set(label, input);
        else state.checked.delete(label);
        renderTags();
      });
    });

    if (btnAddSource && selYear && selExam) {
      btnAddSource.addEventListener('click', () => {
        const year = selYear.value;
        const exam = selExam.value;
        if (!year || !exam) {
          alert('年度と回数を選択してください');
          return;
        }
        const label = `${year}年 ${exam}`;
        if (!state.custom.has(label)) {
          state.custom.add(label);
          renderTags();
        }
        selExam.value = "";
      });
    }

    if (tagInput && suggestionsBox) {
      tagInput.addEventListener('input', (e) => {
        const val = e.target.value.trim();
        suggestionsBox.innerHTML = '';
        if (!val) {
          suggestionsBox.classList.remove('show');
          return;
        }

        const matches = searchDatabase.filter(item => 
          item.label.includes(val) && !item.dom.checked
        );

        if (matches.length > 0) {
          matches.forEach(item => {
            const div = document.createElement('div');
            div.className = 'suggestion-item';
            div.innerHTML = item.label.replace(new RegExp(`(${val})`, 'gi'), '<span class="match">$1</span>');
            div.addEventListener('click', () => {
              item.dom.checked = true;
              item.dom.dispatchEvent(new Event('change'));
              tagInput.value = '';
              suggestionsBox.classList.remove('show');
            });
            suggestionsBox.appendChild(div);
          });
          suggestionsBox.classList.add('show');
        } else {
          suggestionsBox.classList.remove('show');
        }
      });

      document.addEventListener('click', (e) => {
        if (!e.target.closest('.search-wrapper')) suggestionsBox.classList.remove('show');
      });
    }
  },

  /* 7. 生成プロセス (Loading & Result) */
  initGenerationProcess() {
    const generateBtn = document.getElementById('btnStartGeneration');
    const loadingModal = document.getElementById('loadingModal');
    const loadingStep = document.getElementById('loadingStep');
    const resultStep = document.getElementById('resultStep');
    const loadingText = document.getElementById('loadingText');
    const progressBar = document.getElementById('progressBar');
    
    // Result Tabs
    const resultTabs = document.getElementById('resultTabs');
    const resultActions = document.getElementById('resultActions');

    if (!generateBtn || !loadingModal) return;

    // Start Generation
    generateBtn.addEventListener('click', async () => {
      loadingModal.hidden = false;
      loadingStep.style.display = 'flex';
      resultStep.style.display = 'none';
      progressBar.style.width = '0%';
      loadingText.textContent = "準備中...";

      const steps = [
        { text: "問題を生成中...", percent: 25, delay: 800 },
        { text: "図形を描画中...", percent: 50, delay: 1200 },
        { text: "解説を生成中...", percent: 75, delay: 800 },
        { text: "検算中...", percent: 95, delay: 600 }
      ];

      for (const step of steps) {
        loadingText.textContent = step.text;
        progressBar.style.width = `${step.percent}%`;
        await new Promise(r => setTimeout(r, step.delay));
      }

      progressBar.style.width = '100%';
      loadingText.textContent = "完了しました！";
      await new Promise(r => setTimeout(r, 500));

      loadingStep.style.display = 'none';
      resultStep.style.display = 'flex';
      
      // モック結果の注入
      const resProblem = document.getElementById('resProblem');
      const resExplanation = document.getElementById('resExplanation');
      const longText = "\n\n(スクロール確認)\n" + "...\n".repeat(15) + "以上です。";
      if(resProblem) resProblem.textContent = "放物線 y = ax^2 と直線 y = 2x + b が...\n(ここに問題文)" + longText;
      if(resExplanation) resExplanation.textContent = "【解説】\n計算の過程は以下の通りです...\n" + longText;

      // 初期タブを問題にセット
      this.switchResultTab('problem');
    });

    // Tab Switching (Delegation)
    if(resultTabs) {
      resultTabs.addEventListener('click', (e) => {
        const btn = e.target.closest('.result-tab-btn');
        if(!btn) return;
        const targetName = btn.dataset.target;
        this.switchResultTab(targetName);
      });
    }

    // Result Actions (Delegation)
    if(resultActions) {
      resultActions.addEventListener('click', (e) => {
        const btn = e.target.closest('.btn-action');
        if(!btn) return;
        
        // アクションに応じた処理（現状はすべて閉じるだけ）
        loadingModal.hidden = true;
        // Reset view for next time
        loadingStep.style.display = 'flex';
        resultStep.style.display = 'none';
      });
    }
  },

  /* Helper: Switch Result Tab */
  switchResultTab(tabName) {
    const buttons = document.querySelectorAll('.result-tab-btn');
    const sheets = document.querySelectorAll('.a4-sheet');
    
    buttons.forEach(btn => {
      btn.classList.toggle('active', btn.dataset.target === tabName);
    });

    sheets.forEach(sheet => {
      sheet.classList.toggle('active', sheet.id === `sheet-${tabName}`);
    });

    const viewer = document.querySelector('.pdf-viewer-frame');
    if(viewer) viewer.scrollTop = 0;
  }
};