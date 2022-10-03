package service

import (
	"context"
	"crypto/tls"
	mail "email-service/emailModel"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"time"

	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/creator"
	"github.com/unidoc/unipdf/v3/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	gomail "gopkg.in/mail.v2"
)

type Connection struct {
	Server     string
	Database   string
	Collection string
}

const maxUploadSize = 10 * 1024 * 1024 // 10 mb
const dir = "data/download/"

var fileName string

var Collection *mongo.Collection
var ctx = context.TODO()

func (e *Connection) Connect() {
	clientOptions := options.Client().ApplyURI(e.Server)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = license.SetMeteredKey("7bf0687e6377d9e27406c5c6b8f26ad0fcb55cdd74e22275da7c7466cbc3d04f")
	Collection = client.Database(e.Database).Collection(e.Collection)
}

func (e *Connection) EmailWithoutAttachment(mail mail.Mail) (string, error) {

	err := sendMail(mail)
	fmt.Println(err)
	if err != nil {
		return "", err
	}
	fmt.Println("Mail Sent Succefully")
	mail.Time = time.Now()
	insert, err := Collection.InsertOne(ctx, mail)
	fmt.Println(insert)
	if err != nil {
		return "", errors.New("Unable To Insert New Record")
	}
	return "Email Sent Successfully", nil
}

func sendMail(mail mail.Mail) error {
	m := gomail.NewMessage()
	m.SetHeaders(map[string][]string{
		"From":    {m.FormatAddress("mukesh.jangir@gridinfocom.com", "Mukesh")},
		"To":      mail.MailSendTo,
		"Cc":      mail.MailSendCC,
		"Subject": mail.MailSubject,
	})

	m.SetBody("text/plain", mail.MailBody.Salutation+"\n\n"+mail.MailBody.Message+"\n\n"+mail.MailBody.Closing+"\n"+mail.SenderName)

	// Settings for SMTP server
	d := gomail.NewDialer("smtp-relay.sendinblue.com", 587, "vidhi.goel@gridinfocom.com", "pGL756txPrWkSBX4")
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	// Now send E-Mail
	if err := d.DialAndSend(m); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (e *Connection) EmailWithAttachMent(mail mail.Mail, files []*multipart.FileHeader, attachment bool) (string, error) {

	arrayfiles, err := uploadFiles(files)
	if err != nil {
		return "", err
	}
	mail.Attachments = arrayfiles
	err = sendMailWithAttachment(mail, attachment)
	if err != nil {
		return "", err
	}

	fmt.Println("Mail Sent Succefully")
	mail.Time = time.Now()
	_, err = Collection.InsertOne(ctx, mail)

	if err != nil {
		return "", errors.New("Unable To Insert New Record")
	}
	return "Email Sent Successfully", nil
}

func sendMailWithAttachment(mail mail.Mail, attachment bool) error {
	m := gomail.NewMessage()
	m.SetHeaders(map[string][]string{
		"From":    {m.FormatAddress("mukesh.jangir@gridinfocom.com", "Mukesh")},
		"To":      mail.MailSendTo,
		"Cc":      mail.MailSendCC,
		"Subject": mail.MailSubject,
	})

	if attachment {
		for i := range mail.Attachments {
			m.Attach(mail.Attachments[i])
		}
	}

	m.SetBody("text/plain", mail.MailBody.Salutation+"\n\n"+mail.MailBody.Message+"\n\n"+mail.MailBody.Closing+"\n"+mail.SenderName)

	// Settings for SMTP server
	d := gomail.NewDialer("smtp-relay.sendinblue.com", 587, "aarti.kumari@gridinfocom.com", "f7c3OQF8UzC6pIh1")

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	// Now send E-Mail
	if err := d.DialAndSend(m); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func uploadFiles(files []*multipart.FileHeader) ([]string, error) {
	var fileNames []string
	for _, fileHeader := range files {
		fileName = fileHeader.Filename
		fileNames = append(fileNames, dir+fileName)
		if fileHeader.Size > maxUploadSize {
			return fileNames, errors.New("The uploaded file is too big: %s. Please use an file less than 1MB in size: " + fileHeader.Filename)
		}

		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			return fileNames, err
		}

		defer file.Close()

		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			return fileNames, err
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return fileNames, err
		}

		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return fileNames, err
		}

		f, err := os.Create(dir + fileHeader.Filename)
		if err != nil {
			return fileNames, err
		}

		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			return fileNames, err
		}
	}
	return fileNames, nil
}

func (e *Connection) SearchFilter(search mail.Search) ([]*mail.Mail, error) {
	var searchData []*mail.Mail

	filter := bson.D{}

	if search.MailSendTo != "" {
		filter = append(filter, primitive.E{Key: "to", Value: bson.M{"$regex": search.MailSendTo}})
	}
	if search.MailSendCC != "" {
		filter = append(filter, primitive.E{Key: "cc", Value: bson.M{"$regex": search.MailSendCC}})
	}
	if search.MailSendBCC != "" {
		filter = append(filter, primitive.E{Key: "bcc", Value: bson.M{"$regex": search.MailSendBCC}})
	}
	if search.MailSubject != "" {
		filter = append(filter, primitive.E{Key: "subject", Value: bson.M{"$regex": search.MailSubject}})
	}

	t, _ := time.Parse("2006-01-02", search.Date)
	if search.Date != "" {
		filter = append(filter, primitive.E{Key: "time", Value: bson.M{
			"$gte": primitive.NewDateTimeFromTime(t)}})
	}

	result, err := Collection.Find(ctx, filter)

	if err != nil {
		return searchData, err
	}

	for result.Next(ctx) {
		var data mail.Mail
		err := result.Decode(&data)
		if err != nil {
			return searchData, err
		}
		searchData = append(searchData, &data)
	}

	if searchData == nil {
		return searchData, errors.New("No mail found for the given search criteria!")
	}

	return searchData, err
}

func (e *Connection) SearchByEmailId(emailId string) (string, error) {
	var searchResult []*mail.Mail
	os.MkdirAll("data/download", os.ModePerm)
	dir := "data/download/"
	file := "ServiceSearch" + fmt.Sprintf("%v", time.Now().Format("2006-01-02_3_4_5_pm"))

	id, err := primitive.ObjectIDFromHex(emailId)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	cur, err := Collection.Find(ctx, bson.D{primitive.E{Key: "_id", Value: id}})

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	for cur.Next(ctx) {
		var data mail.Mail
		err = cur.Decode(&data)
		if err != nil {
			fmt.Println(err)
			return "", err
		}
		searchResult = append(searchResult, &data)
	}

	log.Println("Pdf")
	_, errPdf := writeDataIntoPDFTable(dir, file, searchResult[0])
	if errPdf != nil {
		fmt.Println(errPdf)
		return "", err
	}

	return "PDF Saved Successfully Having File name: " + file, nil
}

func writeDataIntoPDFTable(dir, file string, data *mail.Mail) (*creator.Creator, error) {

	c := creator.New()
	c.SetPageMargins(20, 20, 20, 20)

	font, err := model.NewStandard14Font(model.HelveticaName)
	if err != nil {
		return c, err
	}

	// Bold font
	fontBold, err := model.NewStandard14Font(model.HelveticaBoldName)
	if err != nil {
		return c, err
	}

	// Generate basic usage chapter.
	if err := basicUsage(c, font, fontBold, data); err != nil {
		return c, err
	}

	err = c.WriteToFile(dir + file + ".pdf")
	if err != nil {
		return c, err
	}
	return c, nil
}

func basicUsage(c *creator.Creator, font, fontBold *model.PdfFont, data *mail.Mail) error {
	// Create chapter.
	ch := c.NewChapter(data.ID.String())
	ch.SetMargins(0, 0, 10, 0)
	ch.GetHeading().SetFont(font)
	ch.GetHeading().SetFontSize(20)
	ch.GetHeading().SetColor(creator.ColorRGBFrom8bit(72, 86, 95))

	contentAlignH(c, ch, font, fontBold, data)

	// Draw chapter.
	if err := c.Draw(ch); err != nil {
		return err
	}

	return nil
}

func contentAlignH(c *creator.Creator, ch *creator.Chapter, font, fontBold *model.PdfFont, data *mail.Mail) {

	// normalFontColorGreen := creator.ColorRGBFrom8bit(4, 79, 3)
	fontSize := 10.0
	to := c.NewParagraph("To" + " :     " + convertArrayOfStringIntoString(data.MailSendTo))
	to.SetFont(font)
	to.SetFontSize(fontSize)
	to.SetMargins(0, 0, 10, 0)
	ch.Add(to)
	cc := c.NewParagraph("Cc" + " :     " + convertArrayOfStringIntoString(data.MailSendCC))
	cc.SetFont(font)
	cc.SetFontSize(fontSize)
	cc.SetMargins(0, 0, 10, 0)
	ch.Add(cc)
	bcc := c.NewParagraph("Bcc" + " :     " + convertArrayOfStringIntoString(data.MailSendBCC))
	bcc.SetFont(font)
	bcc.SetFontSize(fontSize)
	bcc.SetMargins(0, 0, 10, 0)
	ch.Add(bcc)
	sub := c.NewParagraph("Subject" + " :     " + convertArrayOfStringIntoString(data.MailSubject))
	sub.SetFont(font)
	sub.SetFontSize(fontSize)
	sub.SetMargins(0, 0, 10, 0)
	ch.Add(sub)
	salu := c.NewParagraph(data.MailBody.Salutation)
	salu.SetFont(font)
	salu.SetFontSize(fontSize)
	salu.SetMargins(0, 0, 10, 0)
	salu.SetLineHeight(2)
	ch.Add(salu)
	msg := c.NewParagraph(data.MailBody.Message)
	msg.SetFont(font)
	msg.SetFontSize(fontSize)
	msg.SetMargins(0, 0, 10, 0)
	msg.SetLineHeight(2)
	//	a.SetTextAlignment(creator.TextAlignmentJustify)
	ch.Add(msg)
	cls := c.NewParagraph(data.MailBody.Closing)
	cls.SetFont(font)
	cls.SetFontSize(fontSize)
	cls.SetMargins(0, 0, 10, 0)
	cls.SetLineHeight(2)
	//	a.SetTextAlignment(creator.TextAlignmentJustify)
	ch.Add(cls)
}

func convertArrayOfStringIntoString(str []string) string {
	finalData := ""
	y := 0
	for x := range str {
		if y != 0 {
			finalData = finalData + ", "
		}
		finalData = finalData + str[x]
		y++
	}
	y = 0
	return finalData
}
