# PDF Controller

## Description

This controller uses kind pdfDocument to generate a PDF file out of it.

```yaml
apiVersion: k8s.startkubernetes.com.my.domain/v2
kind: PdfDocument
metadata:
  name: pdfdocument
  namespace: default
spec:
  documentName: my-text
  text: |
    ### My document
    Hello **world**!
```

Once applied, it will generate a PDF file with the name `my-text.pdf` in the `/data/pdf` directory of the Job.
