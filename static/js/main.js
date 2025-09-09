new Vue({
    el: '#app',
    data: {
        data: {
            position: 0,
            ezwow: { maxSt: 0 },
            characters: 0,
            cookies: 0,
            races: [],
            classes: [],
            login: ''
        },
        percent: 0,
        findInput: '',
        findOutput: [],
        cooks: { cookie: '' },
        message: ''
    },
    mounted() {
        this.loadStats().then(() => {
            this.loadRaces()
            this.loadClasses()
        })
    },
    methods: {
        loadStats() {
            return axios.get('/api/stats')
                .then(res => {
                    this.data = { ...this.data, ...res.data }
                    this.percent = Math.round((this.data.position / (this.data.ezwow.maxSt || 1)) * 100)
                })
                .catch(err => console.error(err))
        },
        loadRaces() {
            axios.get('/api/races')
                .then(res => {
                    this.data.races = res.data.map(el => ({
                        id: el.ID,
                        value: el.Count,
                        width: Math.round((el.Count / this.data.characters) * 100),
                        name: el.ID
                    }))
                })
                .catch(err => console.error(err))
        },
        loadClasses() {
            axios.get('/api/classes')
                .then(res => {
                    this.data.classes = res.data.map(el => ({
                        id: el.ID,
                        value: el.Count,
                        width: Math.round((el.Count / this.data.characters) * 100),
                        name: el.ID
                    }))
                })
                .catch(err => console.error(err))
        },
        findByName() {
            if (!this.findInput) return;
            axios.get(`/api/characters/${encodeURIComponent(this.findInput)}`)
                .then(res => {
                    this.findOutput = Array.isArray(res.data) ? res.data : [res.data]
                })
                .catch(() => {
                    this.findOutput = []
                    alert('Персонажи не найдены')
                })
        },
        share() {
            alert('Функция отправки куки пока не реализована')
        },
        charsHide() {
            this.findOutput = []
        }
    }
})
