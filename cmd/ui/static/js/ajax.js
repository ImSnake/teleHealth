'use strict';

const authAjax = {

    location: window.location.origin,   //'http://localhost:9999' + ':9999/specs',

    registerRequest(form) {
        console.log('POST register form');
        $.ajax({
            url: authAjax.location + '/register',
            method: 'post',
            data: form.serialize(),
            success: function (data, textStatus, xhr) {
                console.log(xhr.status);
                if (xhr.status === 204) {
                    console.log("данные на сервере");
                    let login = form.find('input[type=email]').val();
                    authorization.signupBox.addClass('hide-element');
                    authorization.authBox.removeClass('hide-element').find('input[type=email]').val(login);
                    authorization.clearFormData(form);
                } else {
                    console.log(data.error);
                }
            },
            error: function (e) {
                console.log(e);
            },
        });
    },

    authRequest(form) {
        console.log('POST authorization form');
        $.ajax({
            url: authAjax.location + '/login',
            method: 'post',
            data: form.serialize(),
            success: function (data, textStatus, xhr) {
                console.log(xhr.status);
                if (xhr.status === 204) {
                    console.log("успешная авторизация");
                    //authorization.headerBox.attr('data-auth', true);
                    //authorization.userUnknown.addClass('hide-element');
                    //authorization.userAuthorized.removeClass('hide-element');
                    //authorization.authBox.addClass('hide-element');
                    //authorization.clearFormData(form);
                    location.reload();
                } else {
                    console.log(data.error);
                }
            },
            error: function (e) {
                console.log(e.responseText);
            },
        });
    },

    logoutRequest() {
        console.log('POST logout request');
        $.ajax({
            url: authAjax.location + '/logout',
            method: 'post',
            success: function (data, textStatus, xhr) {
                console.log(xhr.status);
                if (xhr.status === 204) {
                    (window.location.href.indexOf('/profile') > -1)
                        ? window.location.replace(authAjax.location)
                        : location.reload();
                }
            },
            error: function (e) {
                console.log(e);
            },
        });
    },

    getProfileAvatar(id){
        console.log('GET path to AVATAR image');
        let path = '';
        $.ajax({
            url: authAjax.location + '/get-avatar/' + id,
            method: 'get',
            success: function (data, textStatus, xhr) {
                console.log(xhr.status);
                if (xhr.status === 200) {
                    console.log("ссылка на картинку-аватар получена");
                    path = data.CustomData;
                } else {
                    console.log(data.error);
                }
            },
            error: function (e) {
                console.log(e.responseText);
            },
            complete: function () {
                authorization.renderProfileAvatar(path);
            }
        });
    },
};

const doctorAjax = {

    addDocSpecs(){
        console.log('POST additional specialization');

        let dataSpecs = {
            doctor_id: parseInt(doctor.ID),
            specialization_id: parseInt($('#extra-specialization option:selected').val()),
            certificate_n: $('#extra-document').val(),
            experience: parseInt($('#extra-experience').val())
        };

        $.ajax({
            url: authAjax.location + '/api/doctor/add_spec',
            method: 'post',
            data: JSON.stringify(dataSpecs),
            contentType: "application/json",
            dataType: "json",
            success: function (data, textStatus, xhr) {
                console.log(xhr.status);
                if (xhr.status === 204) {
                    console.log("данные о новой специализации отправлены на сервер");
                    doctor.renderProfSpecs([dataSpecs]);
                } else {
                    console.log(data.error);
                }
            },
            error: function (e) {
                console.log(e);
            },
        });
    },

    addDocBiography(form){
        console.log('POST doctor BIOGRAPHY');

        let dataAbout = {
            biography: form.find('textarea').val(),
        };

        $.ajax({
            url: authAjax.location + '/api/doctor/update_biography/' + doctor.ID,
            method: 'post',
            data: JSON.stringify(dataAbout),
            contentType: "application/json",
            dataType: "json",
            success: function (data, textStatus, xhr) {
                console.log(xhr.status);
                if (xhr.status === 204) {
                    console.log("данные о биографии обновлены");
                } else {
                    console.log(data.error);
                }
            },
            error: function (e) {
                console.log(e);
            },
        });
    },

    addProfileAvatar(){
        let elem = $('#addAvatar');
        var form = elem[0];
        var data = new FormData(form);
        data.append("doctor_id", parseInt(doctor.ID));

        $.ajax({
            type: "POST",
            enctype: 'multipart/form-data',
            url:  authAjax.location + '/api/doctor/add_photo',
            data: data,
            processData: false,
            contentType: false,
            cache: false,
            timeout: 600000,
            success: function (data, textStatus, xhr) {
                console.log(xhr.status);
                console.log("SUCCESS : картинка отправлена на сервер");
            },
            error: function (e) {
                console.log("ERROR : ", e);
            },
            complete: function () {
                authAjax.getProfileAvatar(doctor.ID);
            }
        });
    },

    getSpecsExist() {
        console.log('GET specialisations list');
        $.ajax({
            url: authAjax.location + '/specs_exists',
            method: 'get',
            //dataType: 'json',
            success: function (data, textStatus, xhr) {
                console.log(xhr.status);
                if (xhr.status === 200) {
                    console.log("список специализаций с активными докторами получен");
                    $.each(data, function (key, val) {
                        if (key == "CustomData") {
                            doctor.specsExistList = val;
                        }
                    });
                } else {
                    console.log(data.error);
                }
            },
            error: function (e) {
                console.log(e.responseText);
            },
            complete: function () {
                doctor.checkDoctorData();
            }
        });
    },

    getSpecsList() {
        console.log('GET specialisations list');
        $.ajax({
            url: authAjax.location + '/specs',
            method: 'get',
            //dataType: 'json',
            success: function (data, textStatus, xhr) {
                console.log(xhr.status);
                if (xhr.status === 200) {
                    console.log("список всех специализаций получен");
                    $.each(data, function (key, val) {
                        if (key == "CustomData") {
                            doctor.specsList = val;
                        }
                    });
                } else {
                    console.log(data.error);
                }
            },
            error: function (e) {
                console.log(e.responseText);
            },
            complete: function () {
                doctor.renderSpecsList(doctor.specsList, doctor.specsForm);
            }
        });
    },

    getDoctorsList(id){
        console.log('GET doctors list');
        $.ajax({
            url: authAjax.location + '/specs/' + id,
            method: 'get',
            //dataType: 'json',
            success: function (data, textStatus, xhr) {
                console.log(xhr.status);
                if (xhr.status === 200) {
                    console.log("список докторов заданной специализации получен");
                    console.log(data);
                    $.each(data, function (key, val) {
                        if (key == "CustomData") {
                            doctor.docList = val;
                        }
                    });
                } else {
                    console.log(data.error);
                }
            },
            error: function (e) {
                console.log(e.responseText);
            },
            complete: function () {
                doctor.renderDoctorList();
            }
        });
    },

    getDoctorProfile(id){
        console.log('GET doctor profile');
        $.ajax({
            url: authAjax.location + '/api/doctor/' + id,
            method: 'get',
            dataType: 'json',
            success: function (data, textStatus, xhr) {
                console.log(xhr.status);
                if (xhr.status === 200) {
                    console.log("данные профиля доктора получены");
                    $.each(data, function (key, val) {
                        if (key == "CustomData") {
                            //doctor.docData.push(val);
                            doctor.docData = val;
                        }
                    });
                } else {
                    console.log(data.error);
                }
            },
            error: function (e) {
                console.log(e.responseText);
            },
            complete: function () {
                doctor.defineDocRender();
            }
        });
    },

    getMonthSchedule(string){
        console.log('GET doctor MONTH schedule');
        console.log(string);

        $.ajax({
            url: authAjax.location + '/api/schedule/' + doctor.ID + '/' + string,
            method: 'get',
            success: function (data, textStatus, xhr) {
                console.log(xhr.status);
                if (xhr.status === 200) {
                    console.log("данные о расписании на месяц получены: "  + string);
                    $.each(data, function (key, val) {
                        if (key == "CustomData") {
                            doctor.greenDays  = val;
                        }
                    });
                } else {
                    console.log(data.error);
                }
            },
            error: function (e) {
                console.log(e.responseText);
            },
            complete: function () {
                doctor.renderGreenDays();
            }
        });
    },

    getDaySchedule(string){
        console.log('GET doctor DAY schedule');
        //console.log(string);

        $.ajax({
            url: authAjax.location + '/api/dayschedule/' + doctor.ID + '/' + string,
            method: 'get',
            success: function (data, textStatus, xhr) {
                console.log(xhr.status);
                if (xhr.status === 200) {
                    console.log("данные о расписании на день получены:" + string);
                    doctor.greenHours = data.CustomData;
                } else {
                    console.log(data.error);
                }
            },
            error: function (e) {
                console.log(e.responseText);
            },
            complete: function () {
                doctor.renderGreenHours();
            }
        });
    },

    getMinuteSchedule(string){
        console.log('GET doctor HOUR schedule');
        //console.log(string);

        $.ajax({
            url: authAjax.location + '/api/hourschedule/' + doctor.ID + '/' + string,
            method: 'get',
            success: function (data, textStatus, xhr) {
                console.log(xhr.status);
                if (xhr.status === 200) {
                    console.log("данные о доступных минутах на выбранный час получены:" + string);
                    $.each(data, function (key, val){
                       if(key == "CustomData") {
                           doctor.greenMinutes = val.Gaps;
                       }
                    });
                } else {
                    console.log(data.error);
                }
            },
            error: function (e) {
                console.log(e.responseText);
            },
            complete: function () {
                doctor.renderGreenMinutes();
            }
        });
    },

    saveUserTime(string){
        console.log('POST USER DAY:TIME in doctor schedule');
        //console.log(string);

        $.ajax({
            url: authAjax.location + '/api/enrol/',
            method: 'post',
            data: {
                time: string,
                doctor: parseInt(doctor.ID)
            },
            success: function (data, textStatus, xhr) {
                console.log(xhr.status);
                if (xhr.status === 204) {
                    console.log("запись на прием сохранена на сервере");
                    alert('Вы записаны на прием: ' + string.slice(0, 21));
                } else {
                    console.log(data.error);
                }
            },
            error: function (e) {
                console.log(e.responseText);
            },
            complete: function () {
                doctorAjax.getMinuteSchedule(doctor.getMinuteRequest());
            }
        });
    },

    saveDaySchedule(day, hours){
        console.log('POST SCHEDULE refresh request');
        //console.log(day);
        //console.log(hours);

        $.ajax({
            url: authAjax.location + '/api/makeavailablebyhour/' + day,
            method: 'post',
            data: JSON.stringify(hours),
            contentType: "application/json",
            dataType: "json",
            success: function (data, textStatus, xhr) {
                console.log(xhr.status);
                if (xhr.status === 204) {
                    console.log("расписание на день " + day + " сохранено на сервере");
                } else {
                    console.log(data.error);
                }
            },
            error: function (e) {
                console.log(e.responseText);
            },
            complete: function () {
                doctor.timeTableBox.addClass('hide-element');
                doctorAjax.getMonthSchedule(day.substr(2));
            }
        });
    },
};
