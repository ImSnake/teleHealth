'use strict';

const oftenUsed = {

    // поведение выпадающих по клику элементов
    // показать\скрыть элементы "help" и по-ап скрытых меню, содержащие класс "hover";
    openClickedElem(event, elem) {
        if (event.target.classList.contains('hover')) {
            $(elem).children('.fade').removeClass('hide-element');
        }
        // запуск функции контроля закрытия элемента
        this.onClickClose(elem);
    },

    // закрытие всплывающих элементов (меню, pop-up) при клике вне открытого окна
    onClickClose(elem) { // вызвать в момент показа окна, где elem - окно
        function outsideClickListener(event) {
            if (!elem.contains(event.target) && oftenUsed.isVisible(elem)   // клик не по элементу и элемент виден
                || event.target.classList.contains('close-pop-up')    // клик по "закрыть" внутри открытого блока
                || event.target.tagName === 'A')                     // клик по ссылке внутри открытого блока
            {
                $(elem).children('.fade').addClass('hide-element'); //скрыть
                document.removeEventListener('click', outsideClickListener); //удалить слушатель события
            }
        }

        document.addEventListener('click', outsideClickListener);
    },

    //проверяет открыт ли элемент в момент клика
    isVisible(elem) { //открыто ли условное окно
        return !!elem && !!(elem.offsetWidth || elem.offsetHeight ||
            elem.getClientRects().length);
    },

    // блокировка экрана при всплытии поп-ап
    bodyFix(body) {
        (window.matchMedia('all and (min-width: 630px').matches)
            ? body.css({'overflow': 'hidden', 'padding-right': '0px'})
            : body.css('overflow', 'hidden');
    },

    // снятие блокировки с экрана при закрытии поп-ап
    bodyScroll(body) {
        (window.matchMedia('all and (min-width: 630px').matches)
            ? body.css({'overflow': 'unset', 'padding-right': '0px'})
            : body.css('overflow', 'unset');
    },

    // плавная прокрутка вверх к началу элемента
    scrollToTop(elem) {
        elem.animate({scrollTop: 0}, 'slow');
        //$('html,body').animate({ scrollTop: document.body.scrollHeight }, "slow");
    },

    // плавная прокрутка к вниз до заданного элемента
    scrollToBottom(elem, height) {
        $('html,body').animate({scrollTop: elem.offset().top - height}, 'slow');
    },

    getElemByID(){
        //TODO добавить общий скрипт показа поп-ап по атрибуту data-href
    },

    closePopUpForm(elem){
        elem.parent().parent().parent().addClass('hide-element');
        this.bodyScroll($('body'));
        authorization.clearFormData(elem.siblings('form'));
    },

};

const authorization = {

    headerBox: $('#header-menu'),
    authBox: $('#auth'),
    signupBox: $('#signup'),

    userUnknown: $('#user-unknown'),
    userAuthorized: $('#user-authorized'),
    userAvatar: $('#userAvatar'),

    doctorBox: $('#doctor-data'),
    doctorInput: $('.doctor-data'),

    userPage: $('#userProfile'),
    doctorPage: $('#doctorProfile'),

    // проверка акторизации пользователя
    checkUserAuth() {
        (this.headerBox.attr('data-auth') === 'true')
            ? this.userAuthorized.removeClass('hide-element')
            : this.userUnknown.removeClass('hide-element');

        if (this.headerBox.attr('data-doctor') === 'true'){
            authAjax.getProfileAvatar(this.headerBox.attr('data-id'));
        }
    },

    chooseProfilePage() {
        if (this.headerBox.attr('data-doctor') === 'true') {
            let id = this.headerBox.attr('data-id');
            doctorAjax.getSpecsList();
            doctorAjax.getDoctorProfile(id);
            this.doctorPage.removeClass('hide-element');
        } else {
            this.userPage.removeClass('hide-element');
        }
    },

    // сравнить пароли
    passwordCheck(first, second, message) {
        (first.val() == second.val())
            ? message.html('')
            : message.html('Пароли не совпадают');
    },

    // если отмечен чекбокс "Я - Доктор": открыть форму для регистрации доктора + отметить все поля формы обязательными для заполнения
    fromCheckboxAction() {
        this.doctorBox.toggleClass('hide-element');
        (this.doctorBox.hasClass('hide-element') === false)
            ? $.each(this.doctorInput, function () {
                $(this).attr('required', true)
            })
            : $.each(this.doctorInput, function () {
                $(this).attr('required', false)
            });
    },

    // очистить заполненные поля формы
    clearFormData(form) {
        form.find('input').val('');
        if (form.parent().hasClass('register')) {
            this.doctorBox.removeClass('hide-element');
            form.find('input[type=checkbox]').prop('checked', false);
            this.fromCheckboxAction();
        }
    },

    renderProfileAvatar(path){
        console.log(path);
        (path !== '')
            ? this.userAvatar.attr('src', path)
            : this.userAvatar.attr('src', '/static/images/avatars/avatar.png');

        if(path !== '' && $('#profilePage').length > 0){
            $('.doc_avatar img').attr('src', path);
        }
    },
};

const doctor = {

    specsAction: $('#specs-action'),
    specsForm: $('#specialization'),
    specsHeading: $('#specsHeading'),
    specsExistList: null,
    specsList: null,
    specsCurrent: null,

    docBox: $('#doctorsList'),
    docBoxTmpl: $('.catalog__block.template.hide-element'),
    docReviewTmpl: undefined,
    docList: null,
    docData: null,

    docPage: $('#doctorPage'),
    docProf: authorization.doctorPage,
    ID: null,

    calendarBox: $('#myCalendar'),
    timeTableBox: $('#timetable'),
    timeHour: $('#timeHour'),
    timeMin: $('#timeMinute'),
    timeBtn: $('#timeConfirm'),
    day: null,
    month: null,
    year: null,
    //hourCheck: null,
    greenDays: null,
    greenHours: null,
    greenMinutes: null,

    pickDate: null,
    pickHour: null,
    pickMin: null,


    renderSpecsList(list, box) {
        let items = [];
        $.each(list, function (key, val) {
            items.push('<option value="' + val.ID + '">' + val.spec_name + '</option>');
        });
        box.append(items);

        $(this.specsAction).on('change', function () {
            let id = $(this).find(":selected").val();
            window.location.replace(window.location.origin + '/catalog/' + id);
        });
    },

    checkDoctorData(){
        // если это страница КАТАЛОГ, то запросить список докторов
        if (this.docBox.length > 0) {
            doctor.getCurrentSpecs();
        }
        // если это страница ДОКТОР, то запросить данные доктора
        if (this.docPage.length > 0) {
            doctor.getSpecsAndDocID();
        }
        this.renderSpecsList(this.specsExistList, this.specsAction);
    },

    getCurrentSpecs() {
        let url = window.location.href.split('/');
        let id = url[url.length - 1];
        this.specsCurrent = id;
        doctorAjax.getDoctorsList(id);
    },

    getSpecsAndDocID() {
        this.specsCurrent = window.location.hash.substr(1);
        let url = window.location.href;
        let string = url.slice(0, url.lastIndexOf("#"));
        string = string.split('/');
        let docId = string[string.length - 1];
        doctorAjax.getDoctorProfile(docId);
    },

    renderSpecsTitle() {
        console.log('добавить заголовок');
        $.each(doctor.specsExistList, function (key, val) {
            if (parseInt(val.ID) === parseInt(doctor.specsCurrent)) {
                doctor.specsHeading.text(val.spec_name);
            }
        });
    },

    renderDoctorList() {

        if (this.docList != null) {
            this.docList.forEach(function (val) {
                let template = doctor.docBoxTmpl.clone();
                template.removeClass('template hide-element');
                if (val.PhotoURL !== '') {
                    template.find('.cat_doc_avatar img').attr('src', val.PhotoURL);
                }
                template.find('.cat_doc_id').attr('href', '/doctor/' + val.ID + '#' + doctor.specsCurrent)
                template.find('.cat_doc_name').text(val.Name);
                template.find('.cat_doc_patronymic').text(val.Patronymic);
                template.find('.cat_doc_surname').text(val.Surname);
                template.find('.cat_doc_exp').text(val.Experience);
                template.find('.cat_doc_reviews').text(val.CountReview);
                doctor.getRatingHtml(val.Rating, template.find('.cat_doc_rating'));
                doctor.docBox.append(template);
            });
        } else {
            doctor.docBox.append("<div>Пока нет ни одного доктора</div>");
        }
        this.renderSpecsTitle();
    },

    defineDocRender() {
        if (this.docPage.length > 0) {
            console.log('это страница ДОКТОРА');
            this.renderDoctorPage();
        } else {
            console.log('это профиль ДОКТОРА');
            this.renderDoctorProfile();
        }

        // провекра на необходимость вывода календаря
        if (this.calendarBox.length > 0) {
            this.addCalendar();
            let text = this.calendarBox.find('.month-head').text();
            let string = text.substr(4, 3) + text.slice(-4);
            doctorAjax.getMonthSchedule(string);
        }
    },

    renderDoctorPage() {

        console.log(this.docData);

        this.renderSpecsTitle();

        $.each(doctor.docData, function (key, val) {

            if(key === 'DoctorData'){

                doctor.ID = val.ID; // сохранить ID доктора

                doctor.docPage.find('.doc_surname').text(val.surname);
                doctor.docPage.find('.doc_name').text(val.name);
                doctor.docPage.find('.doc_patronymic').text(val.patronymic);
                doctor.docPage.find('.doc_about').text(val.biography);
                if (val.photo_url != null) {
                    doctor.docPage.find('.doc_avatar img').attr('src', val.photo_url);
                } else {
                    doctor.docPage.find('.doc_avatar img').attr('src', '/static/images/avatars/avatar-big.png');
                }
            }
            if(key === 'Age'){
                doctor.docPage.find('.doc_age span').text(val);
            }
            if(key === 'Rating'){
                doctor.getRatingHtml(val,  doctor.docPage.find('.doc_rating'));
            }
            if(key === 'ReviewsData'){
                let count = 0;
                $.each(val, function (){
                     count += 1;
                });
                doctor.docPage.find('.reviews__count').text(count);
            }
            if(key === 'Specializations'){
                $.each(val, function (key, val){
                    if(parseInt(val.specialization_id) === parseInt(doctor.specsCurrent)){
                        doctor.docPage.find('.doc_exp span').text(val.experience);
                    }
                });
            }
        });

        this.docPage.removeClass('hide-element');

//****** слушатели событий на календарь ********************************************************************************

        this.calendarBox.on('click', 'td', function (){
            if(!$(this).hasClass('prevDate') && $(this).hasClass('available')){
                doctor.timeTableBox.find('.timetable__heading__text').text( doctor.pickDate.toString().substr(4, 12));
                doctor.timeTableBox.removeClass('hide-element').find('.timetable__hour').removeClass('active');
                doctor.timeMin.addClass('hide-element');
                doctor.timeBtn.addClass('hide-element');
                doctorAjax.getDaySchedule(doctor.getHourRequest());
            } else {
                doctor.timeTableBox.addClass('hide-element');
            }
        });

        this.timeHour.on('click', '.timetable__hour', function (){

            doctor.timeHour.find('.timetable__hour').removeClass('active');
            doctor.pickHour = $(this).text().slice(0, -2);

            $.each(doctor.timeMin.children('div'), function (){
                $(this).text(doctor.pickHour + $(this).text().slice(-2));
            });

            if(!$(this).hasClass('blocked')){
                $(this).addClass('active');
                doctor.timeMin.removeClass('hide-element').children('div').removeClass('active');
                doctor.timeBtn.addClass('hide-element');

                //let text = doctor.pickDate.toString();
                //let string = text.substr(8, 2) + text.substr(4, 3) + text.substr(11, 4) + 'T' + doctor.pickHour;
                doctorAjax.getMinuteSchedule(doctor.getMinuteRequest());
            } else {
                doctor.timeMin.addClass('hide-element').children('div').removeClass('active');
            }
        });

        this.timeMin.on('click', '.schedule__minutes__item', function (){
            doctor.timeMin.children('div').removeClass('active');

            if(!$(this).hasClass('blocked')){
                doctor.pickMin = $(this).addClass('active').text().slice(-2);
                doctor.timeBtn.removeClass('hide-element').find('span').text($(this).attr('data-time').slice(0, -7));
            } else {
                doctor.timeBtn.addClass('hide-element').find('span').text();
            }

        });

        $('#timeConfirm button').on('click', function (){
            if (authorization.headerBox.attr('data-auth') === 'false'){
                alert('Пользователь не авторизован');
/*            } else if (authorization.headerBox.attr('data-doctor') === 'true') {
                alert('Авторизованный пользователь  - доктор, а не пациент.\n' +
                    'Невозможно записаться к другому доктору или к самому себе.');*/
            } else {
                let data = doctor.pickDate.toString().substr(0, 16) + doctor.timeMin.find('.active.available').attr('data-time').slice(17, -4) + ' GMT+0000'//doctor.pickDate.toString().substr(24, 9 );
                doctorAjax.saveUserTime(data);
               }
        });
    },

    renderDoctorProfile() {

        console.log(this.docData);

        $.each(doctor.docData, function (key, val) {

            if(key === 'DoctorData'){

                doctor.ID = val.ID; // сохранить ID доктора

                if (val.photo_url != null) {
                    doctor.docProf.find('.doc_avatar img').attr('src', val.photo_url);
                } else {
                    doctor.docProf.find('.doc_avatar img').attr('src', '/static/images/avatars/avatar-big.png');
                }
                doctor.docProf.find('.doc_created_at span').text(val.created_at.slice(0, -10));
                doctor.docProf.find('.doc_name').text(val.name);
                doctor.docProf.find('.doc_patronymic').text(val.patronymic);
                doctor.docProf.find('.doc_surname').text(val.surname);
                doctor.docProf.find('.doc_birthday span').text(val.birthdate.slice(0, -10));
                doctor.docProf.find('.doc_about textarea').text(val.biography);
            }

            if(key === 'Age'){
                doctor.docPage.find('.doc_age span').text(val);
            }

            if(key === 'Rating'){
                doctor.getRatingHtml(val,  doctor.docProf.find('.doc_rating'));
            }

            if(key === 'ReviewsData'){
                let count = 0;
                $.each(val, function (){
                    count += 1;
                });
                doctor.docProf.find('.reviews__count').text(count);
            }

            if(key === 'Specializations'){
                doctor.renderProfSpecs(val);
            }
        });

        this.renderSpecsList(this.specsList, $('#extra-specialization'));

//****** слушатели событий на календарь ********************************************************************************

        this.calendarBox.on('click', 'td', function (){
            if(!$(this).hasClass('prevDate')){
                doctor.timeTableBox.find('.timetable__heading__text').text(doctor.pickDate.toString().substr(4, 12));
                doctor.timeTableBox.removeClass('hide-element').find('.timetable__hour').removeClass('active');
                doctorAjax.getDaySchedule(doctor.getHourRequest());
            } else {
                doctor.timeTableBox.addClass('hide-element');
            }
        });

        this.timeHour.on('click', '.timetable__hour, button', function (){
            if($(this).hasClass('timetable__hour')) {
                $(this).toggleClass('active').removeClass('available');
            }
            if($(this).hasClass('check-all')){
                doctor.timeHour.find('.timetable__hour').addClass('active').removeClass('available');
            }
            if($(this).hasClass('delete-all')){
                doctor.timeHour.find('.timetable__hour').removeClass('active').removeClass('available');
            }
            if($(this).hasClass('refresh')){
                doctor.getAllCheckedHours();
            }
        });

        $('#add-specs form').on('submit', function (e){
           e.preventDefault();
           doctorAjax.addDocSpecs();
           oftenUsed.closePopUpForm($(this));
        });

        $('#add-avatar form').on('submit', function (e){
            e.preventDefault();
            doctorAjax.addProfileAvatar();
            oftenUsed.closePopUpForm($(this));
        });

        $('#docBiography').on('submit', function (e){
            e.preventDefault();
            doctorAjax.addDocBiography($(this));
        });
    },

    renderReviews() {

        if (this.docData.ReviewsData != null && $('#reviews .review__box').length === 1) {
            this.docReviewTmpl = $('.review__box.template.hide-element');
            $.each(this.docData.ReviewsData, function (key, val){
                let template = doctor.docReviewTmpl.clone();
                template.removeClass('template hide-element');
                template.find('.review__data').text(val.created_at.slice(0, -10));
                template.find('.review__text').text(val.text);
                doctor.getRatingHtml(val.rating, template.find('.review__rating'));
                $('#reviews .pop-up').append(template);
            });
            return;
        }

        if (this.docData.ReviewsData === null) {
            $('#reviews .pop-up').append('<div>Нет отзывов</div>');
        }
    },

    // визуализировать рейтинг врача
    getRatingHtml(num, elem) {

        elem.attr('data-rating', num);

        if (parseFloat(num) === 0) {
            return elem.append(
                '<span class="star-estimation-empty"></span>\n' +
                ' <span class="star-estimation-empty"></span>\n' +
                ' <span class="star-estimation-empty"></span>\n' +
                ' <span class="star-estimation-empty"></span>\n' +
                ' <span class="star-estimation-empty"></span>');
        }
        if (parseFloat(num) >= 4.75) {
            return elem.append(
                '<span class="star-estimation"></span>\n' +
                ' <span class="star-estimation"></span>\n' +
                ' <span class="star-estimation"></span>\n' +
                ' <span class="star-estimation"></span>\n' +
                ' <span class="star-estimation"></span>');
        }
        if (parseFloat(num) >= 4.5) {
            return elem.append(
                '<span class="star-estimation"></span>\n' +
                ' <span class="star-estimation"></span>\n' +
                ' <span class="star-estimation"></span>\n' +
                ' <span class="star-estimation"></span>\n' +
                ' <span class="star-estimation-half"></span>');
        }
        if (parseFloat(num) >= 3.8) {
            return elem.append(
                '<span class="star-estimation"></span>\n' +
                ' <span class="star-estimation"></span>\n' +
                ' <span class="star-estimation"></span>\n' +
                ' <span class="star-estimation"></span>\n' +
                ' <span class="star-estimation-empty"></span>');
        }
        if (parseFloat(num) >= 3.5) {
            return elem.append(
                '<span class="star-estimation"></span>\n' +
                ' <span class="star-estimation"></span>\n' +
                ' <span class="star-estimation"></span>\n' +
                ' <span class="star-estimation-half"></span>\n' +
                ' <span class="star-estimation-empty"></span>');
        }
        if (parseFloat(num) >= 2.8) {
            return elem.append(
                '<span class="star-estimation"></span>\n' +
                ' <span class="star-estimation"></span>\n' +
                ' <span class="star-estimation"></span>\n' +
                ' <span class="star-estimation-empty"></span>\n' +
                ' <span class="star-estimation-empty"></span>');
        }
        if (parseFloat(num) >= 2) {
            return elem.append(
                '<span class="star-estimation"></span>\n' +
                ' <span class="star-estimation"></span>\n' +
                ' <span class="star-estimation-half"></span>\n' +
                ' <span class="star-estimation-empty"></span>\n' +
                ' <span class="star-estimation-empty"></span>');
        }
        if (parseFloat(num) >= 1.3) {
            return elem.append(
                '<span class="star-estimation"></span>\n' +
                ' <span class="star-estimation"></span>\n' +
                ' <span class="star-estimation-empty"></span>\n' +
                ' <span class="star-estimation-empty"></span>\n' +
                ' <span class="star-estimation-empty"></span>');
        }
        if (parseFloat(num) >= 0.5) {
            return elem.append(
                '<span class="star-estimation"></span>\n' +
                ' <span class="star-estimation-empty"></span>\n' +
                ' <span class="star-estimation-empty"></span>\n' +
                ' <span class="star-estimation-empty"></span>\n' +
                ' <span class="star-estimation-empty"></span>');
        }
        if (parseFloat(num) < 0.5) {
            return elem.append(
                '<span class="star-estimation-half"></span>\n' +
                ' <span class="star-estimation-empty"></span>\n' +
                ' <span class="star-estimation-empty"></span>\n' +
                ' <span class="star-estimation-empty"></span>\n' +
                ' <span class="star-estimation-empty"></span>');
        }
    },

    renderProfSpecs(arr){
        let specs = arr;
        let template = doctor.docProf.find('.specs_info.template.hide-element');
        $.each(specs, function (key, val){
            let specsBox = template.clone();
            let id = val.specialization_id;
            let name = '';
            $.each(doctor.specsList, function (key, val){
                if(parseInt(val.ID) === parseInt(id)){
                    name = val.spec_name;
                }
            });
            specsBox.find('.specs_name').text(name);
            specsBox.find('.specs_exp span').text(val.experience);
            specsBox.find('.specs_sert span').text(val.certificate_n);
            doctor.docProf.find('.doc_specs').append(specsBox.removeClass('template hide-element'));
        });
    },

    // добавить на страницу календарь
    addCalendar() {
        this.calendarBox.calendar({
            date: new Date(),
            autoSelect: true,
            select: function (date) {
                doctor.pickDate = date;
            }
        });
        this.setPrevDays();
    },

    // отметить на календаре прошедшие даты
    setPrevDays() {
        this.day = new Date().getUTCDate();
        this.month = new Date().getUTCMonth() + 1;
        this.year = new Date().getUTCFullYear();

        this.compareDates();

        $('.myCalendar').on('click', '.month-head button', function () {
            let elem = $(this);
            if (elem.hasClass('ic-arrow-angle-left') || elem.hasClass('ic-arrow-angle-right')) {
                doctor.compareDates();
                let string = elem.parent().text().substr(0, 3) + elem.parent().text().slice(-4);
                doctorAjax.getMonthSchedule(string);
            }
        });
    },

    // при перелистывании кадендаря - сравнить выбранные даты с текущей
    compareDates() {

        let xMonth = $('.myCalendar.nao-month .year-body td.active').attr('data-month');
        let xYear = $('.myCalendar .month-head div').html();
        let xDays = $('.myCalendar.nao-month .month-body .month-days td');

        if (this.year >= xYear) {

            if (xYear == this.year && xMonth == this.month) {
                $.each(xDays, function () {
                    if ($(this).text() < doctor.day) {
                        $(this).addClass('prevDate');
                    }
                });
            }
            if (xYear < this.year || xMonth < this.month) {
                $.each(xDays, function () {
                    $(this).addClass('prevDate');
                });
            }
        }
    },

    renderGreenDays(){
      if (this.greenDays !== null) {
          $.each(this.greenDays, function (key, val){
              $.each($('table.month-body td'), function (){
                  if(!$(this).hasClass('prevDate') && parseInt($(this).text()) === parseInt(key)){
                      $(this).addClass('available');
                  }
              });
          });
      }

    },

    renderGreenHours(){
        if (this.greenHours !== null) {
            let arr = [];
            $.each(this.greenHours, function (key, val){
                let time = key;
                if (key.length === 1){
                    time = '0' + key;
                }
                arr.push(time + ':00');
            });
            if($('#profilePage').length > 0) {
                $.each(doctor.timeHour.find('.timetable__hour'), function (){
                    let time = $(this).removeClass('available');
                    arr.forEach(function(val) {
                        if (time.text() === val){
                            time.addClass('available');
                        }
                    });
                });
            } else {
                $.each(doctor.timeHour.find('.timetable__hour'), function (){
                    let time = $(this).removeClass('available').addClass('blocked');
                    arr.forEach(function(val) {
                        if (time.text() === val){
                            time.addClass('available').removeClass('blocked');
                        }
                    });
                });
            }
        }
    },

    renderGreenMinutes(){
        let arr = []
        $.each(this.greenMinutes, function (key, val){
            if (val.Available === true){
                let time = key.substr(17, 5);
                arr.push({time: time, data: key});
            }
        });
        console.log(arr);
        $.each(doctor.timeMin.find('.schedule__minutes__item'), function (){
            let time = $(this).removeClass('available active').addClass('blocked').attr('data-time', '');
            arr.forEach(function(key, val) {
                if (time.text() == key.time){
                    time.addClass('available').removeClass('blocked').attr('data-time', key.data);
                }
            });
        });
    },

    getHourRequest() {
        let text = doctor.pickDate.toString();
        let string = text.substr(8, 2) + text.substr(4, 3) + text.substr(11, 4);
        return string;
    },

    getMinuteRequest(){
        let text = doctor.pickDate.toString();
        let string = text.substr(8, 2) + text.substr(4, 3) + text.substr(11, 4) + 'T' + doctor.pickHour;
        return string.slice(0, -1);
    },

    getAllCheckedHours(){
        let arr = [];
        $.each(this.timeHour.find('.active'), function (){
            let hour = $(this).text().slice(0, -3);
            arr.push(parseInt(hour));
        });
        doctorAjax.saveDaySchedule(this.getHourRequest(), arr);
    },
};

(function ($) {
    $(document).ready(
        function () {

// ===== O N  L O A D =========================================================================================

            // проверка авторизации пользователя
            authorization.checkUserAuth();

            // получение списка специализаций
            if ($('#specsBox').length > 0) {
                doctorAjax.getSpecsExist();
            }

            // страница /profile, определить показ профиля: USER or DOCTOR
            if ($('#profilePage').length > 0) {
                authorization.chooseProfilePage();
            }


// ===== C O M M O N ==========================================================================================

            //открыть / скрыть всплывающие элементы и поп-апы
            $('.hover').on('click', function (event) {
                oftenUsed.openClickedElem(event, this);
            });

// ===== H E A D E R / F O O T E R ===========================================================================

            //поведение выпадающего меню в header для user
            $('.header__user div').on('click', (function () {
                $(this).children('.fade').removeClass('hide-element');
            })).on('mouseleave', function () {
                $(this).children('.fade').addClass('hide-element');
            });

            // показать поп-ап с ID атрибута "data-href" переданного элемента
            $('#header-menu .menu-list span, .add-avatar button, .new_specs button, .doc_reviews').on('click', function () {
                let elem = $(this).attr('data-href');
                $(elem).removeClass('hide-element');
                oftenUsed.bodyFix($('body'));
                if (elem === '#signup') {
                    doctorAjax.getSpecsList();
                }
                if(elem === "#reviews"){
                    doctor.renderReviews();
                }
            });

            //закрыть поп-ап регистрации / авторизации / загрузки аватара / добавления специализации
            $('.authorization .close-pop-up').on('click', function () {
                oftenUsed.closePopUpForm($(this));
            });

            //установить текущий год в подвале
            $("#year").text(new Date().getFullYear());

// ===== A U T H O R I Z A T I O N   E V E N T   L I S T E N E R S =====================================================

            // регистрация: запуск функции проверки пароля
            $('#add-password, #repeat-password').on('keyup', this, function () {
                authorization.passwordCheck($("#add-password"), $("#repeat-password"), $("#pass-message"));
            });

            //  регистрация: поведение чекбокса "Я - Доктор"
            $('#doctor-true').on('change', function () {
                authorization.fromCheckboxAction();
            });

            //  регистрация: отправка формы
            $('#signup form').on('submit', function (e) {
                e.preventDefault();
                ($('#pass-message').text() === '' && $("#add-password").val().length >= 6)   //check password
                    ? authAjax.registerRequest($("#signup form"))
                    : alert("проблема с паролем");
            });

            //  авторизация: отправка формы
            $('#auth form').on('submit', function (e) {
                e.preventDefault();
                authAjax.authRequest($("#auth form")); //"checking password"
            });

            // логаут
            $('#logout').on('click', function () {
                authAjax.logoutRequest();
            });


// ==========================================================================================================




        });
})
(jQuery);


