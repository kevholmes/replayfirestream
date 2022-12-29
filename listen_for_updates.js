// Listen for updates on a local firestore emulator

const Firestore = require('@google-cloud/firestore');

process.env['FIRESTORE_EMULATOR_HOST'] = 'localhost:8070';

const db = new Firestore({
  projectId: 'localhost',
});

async function listenForStatusUpdates() {
    //const observer = await db.collection(`/account/200/vehicle/1337/report_data`).where('type', '==', 'status')
    const observer = await db.collectionGroup(`report_data`).where('type', '==', 'status')
    .onSnapshot(querySnapshot => {
        querySnapshot.docChanges().forEach(change => {
            if (change.type === 'added') {
                var reportData = change.doc.data();
                var thenSec = reportData.reportTimestamp._seconds
                //console.log('thenSec: ', thenSec)
                var thenNano = reportData.reportTimestamp._nanoseconds
                //console.log('thenNano: ', thenNano)
                // convert thenSec to millis, add to it our nanoseconds value converted to milliseconds
                var then = (thenSec * 1000) + Math.round(thenNano / 1000000)
                console.log('then: ', then)
                var now = Date.now(); // milliseconds in javascript
                console.log('now : ', now)
                const diffTime = now - then
                console.log('diff time milliseconds: ', diffTime);
                const diffSec = (diffTime / 1000);
                console.log('diff time seconds     : ', Math.round(diffSec));
                //console.log('status point added: ', change.doc.data());
            }
        });
    }, err => {
        console.log(`Encountered error: ${err}`)
    });
}
listenForStatusUpdates();
